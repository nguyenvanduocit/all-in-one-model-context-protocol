package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func RegisterYouTubeChannelTools(s *server.MCPServer) {
	// Update video tool
	updateVideoTool := mcp.NewTool("youtube_update_video",
		mcp.WithDescription("Update a video's title and description on YouTube"),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("ID of the video to update")),
		mcp.WithString("title", mcp.Required(), mcp.Description("New title of the video")),
		mcp.WithString("description", mcp.Required(), mcp.Description("New description of the video")),
		mcp.WithString("keywords", mcp.Required(), mcp.Description("Comma-separated list of keywords for the video")),
		mcp.WithString("category", mcp.Required(), mcp.Description("Category ID for the video. See https://developers.google.com/youtube/v3/docs/videoCategories/list for more information.")),
	)
	s.AddTool(updateVideoTool, util.ErrorGuard(youtubeUpdateVideoHandler))

	getVideoDetailsTool := mcp.NewTool("youtube_get_video_details",
		mcp.WithDescription("Get details (title, description, ...) for a specific video"),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("ID of the video")),
	)
	s.AddTool(getVideoDetailsTool, util.ErrorGuard(youtubeGetVideoDetailsHandler))

	// List my channels tool
	listMyChannelsTool := mcp.NewTool("youtube_list_videos",
		mcp.WithDescription("List YouTube videos managed by the user"),
		mcp.WithString("channel_id", mcp.Required(), mcp.Description("ID of the channel to list videos for")),
		mcp.WithNumber("max_results", mcp.Required(), mcp.Description("Maximum number of videos to return")),
	)
	s.AddTool(listMyChannelsTool, util.ErrorGuard(youtubeListVideosHandler))

}

var youtubeService = sync.OnceValue[*youtube.Service](func() *youtube.Service {
	ctx := context.Background()

	tokenFile := os.Getenv("GOOGLE_TOKEN_FILE")
	if tokenFile == "" {
		panic("GOOGLE_TOKEN_FILE environment variable must be set")
	}

	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		panic("GOOGLE_CREDENTIALS_FILE environment variable must be set")
	}

	client := services.GoogleHttpClient(tokenFile, credentialsFile)

	srv, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		panic(fmt.Sprintf("failed to create YouTube service: %v", err))
	}

	return srv
})

func youtubeUpdateVideoHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	var videoID string
	if videoIDArg, ok := arguments["video_id"]; ok {
		videoID = videoIDArg.(string)
	} else {
		return mcp.NewToolResultError("video_id is required"), nil
	}

	var title string
	if titleArg, ok := arguments["title"]; ok {
		title = titleArg.(string)
	}

	var description string
	if descArg, ok := arguments["description"]; ok {
		description = descArg.(string)
	}

	var keywords string
	if keywordsArg, ok := arguments["keywords"]; ok {
		keywords = keywordsArg.(string)
	}

	var category string
	if categoryArg, ok := arguments["category"]; ok {
		category = categoryArg.(string)
	}

	updateCall := youtubeService().Videos.Update([]string{"snippet"}, &youtube.Video{
		Id: videoID,
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			Tags:        strings.Split(keywords, ","),
			CategoryId:  category,
		},
	})

	_, err := updateCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update video: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully updated video with ID: %s", videoID)), nil
}


func youtubeGetVideoDetailsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	videoID, ok := arguments["video_id"].(string)
	if !ok {
		return mcp.NewToolResultError("video_id is required"), nil
	}

	listCall := youtubeService().Videos.List([]string{"snippet", "contentDetails", "statistics"}).
		Id(videoID)
	listResponse, err := listCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get video details: %v", err)), nil
	}

	if len(listResponse.Items) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("video with ID %s not found", videoID)), nil
	}

	video := listResponse.Items[0]
	result := fmt.Sprintf("Title: %s\n", video.Snippet.Title)
	result += fmt.Sprintf("Description: %s\n", video.Snippet.Description)
	result += fmt.Sprintf("Video ID: %s\n", video.Id)
	result += fmt.Sprintf("Duration: %s\n", video.ContentDetails.Duration)
	result += fmt.Sprintf("Views: %d\n", video.Statistics.ViewCount)
	result += fmt.Sprintf("Likes: %d\n", video.Statistics.LikeCount)
	result += fmt.Sprintf("Comments: %d\n", video.Statistics.CommentCount)

	return mcp.NewToolResultText(result), nil
}

func youtubeListVideosHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	var channelID string
	if channelIDArg, ok := arguments["channel_id"]; ok {
		channelID = channelIDArg.(string)
	} else {
		return mcp.NewToolResultError("channel_id is required"), nil
	}

	var maxResults int64
	if maxResultsArg, ok := arguments["max_results"].(float64); ok {
		maxResults = int64(maxResultsArg)
	} else {
		maxResults = 10
	}

	// Get the channel's uploads playlist ID
	channelsListCall := youtubeService().Channels.List([]string{"contentDetails"}).
		Id(channelID)
	channelsListResponse, err := channelsListCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get channel details: %v", err)), nil
	}

	if len(channelsListResponse.Items) == 0 {
		return mcp.NewToolResultError("channel not found"), nil
	}

	uploadsPlaylistID := channelsListResponse.Items[0].ContentDetails.RelatedPlaylists.Uploads

	// List videos in the uploads playlist
	playlistItemsListCall := youtubeService().PlaylistItems.List([]string{"snippet"}).
		PlaylistId(uploadsPlaylistID).
		MaxResults(maxResults)
	playlistItemsListResponse, err := playlistItemsListCall.Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list videos: %v", err)), nil
	}

	var result string
	for _, playlistItem := range playlistItemsListResponse.Items {
		videoID := playlistItem.Snippet.ResourceId.VideoId
		videoDetailsCall := youtubeService().Videos.List([]string{"snippet", "statistics"}).
			Id(videoID)
		videoDetailsResponse, err := videoDetailsCall.Do()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get video details: %v", err)), nil
		}

		if len(videoDetailsResponse.Items) > 0 {
			video := videoDetailsResponse.Items[0]
			result += fmt.Sprintf("Video ID: %s\n", video.Id)
			result += fmt.Sprintf("Published At: %s\n", video.Snippet.PublishedAt)
			result += fmt.Sprintf("View Count: %d\n", video.Statistics.ViewCount)
			result += fmt.Sprintf("Like Count: %d\n", video.Statistics.LikeCount)
			result += fmt.Sprintf("Comment Count: %d\n", video.Statistics.CommentCount)
			result += fmt.Sprintf("Title: %s\n", video.Snippet.Title)
			result += fmt.Sprintf("Description: %s\n", video.Snippet.Description)
			result += "-------------------\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}