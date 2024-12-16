package tools

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
)

const (
	RE_YOUTUBE        = `(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})`
	USER_AGENT        = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36`
	RE_XML_TRANSCRIPT = `<text start="([^"]*)" dur="([^"]*)">([^<]*)<\/text>`
)

// RegisterYouTubeTool registers the YouTube transcript tool with the MCP server
func RegisterYouTubeTool(s *server.MCPServer) {
	tool := mcp.NewTool("youtube_transcript",
		mcp.WithDescription("Get YouTube video transcript"),
		mcp.WithString("video_id", mcp.Required(), mcp.Description("YouTube video ID")),
	)

	s.AddTool(tool, util.ErrorGuard(youtubeTranscriptHandler))
}

func youtubeTranscriptHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Get URL from arguments
	videoID, ok := arguments["video_id"].(string)
	if !ok {
		return nil, fmt.Errorf("video_id argument is required")
	}

	// Fetch transcript
	transcripts, videoTitle, err := FetchTranscript(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transcript: %v", err)
	}

	// Build result string
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Title: %s\n\n", videoTitle))

	for _, transcript := range transcripts {
		// Decode HTML entities in the text
		decodedText := decodeHTML(transcript.Text)
		// Format timestamp in [HH:MM:SS] format
		timestamp := formatTimestamp(transcript.Offset)

		builder.WriteString(timestamp)
		builder.WriteString(decodedText)
		builder.WriteString("\n")
	}

	return mcp.NewToolResultText(builder.String()), nil
}

// Error types
type YoutubeTranscriptError struct {
	Message string
}

func (e *YoutubeTranscriptError) Error() string {
	return fmt.Sprintf("[YoutubeTranscript] 🚨 %s", e.Message)
}

type TranscriptResponse struct {
	Text     string
	Duration float64
	Offset   float64
	Lang     string
}


// FetchTranscript retrieves the transcript for a YouTube video
func FetchTranscript(videoId string) ([]TranscriptResponse, string, error) {
	identifier, err := retrieveVideoId(videoId)
	if err != nil {
		return nil, "", err
	}

	videoPageURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", identifier)

	videoPageResponse, err := services.DefaultHttpClient().Get(videoPageURL)
	if err != nil {
		return nil, "", err
	}
	defer videoPageResponse.Body.Close()

	videoPageBody, err := io.ReadAll(videoPageResponse.Body)
	if err != nil {
		return nil, "", err
	}

	// Extract video title
	titleRegex := regexp.MustCompile(`<title>(.+?) - YouTube</title>`)
	titleMatch := titleRegex.FindSubmatch(videoPageBody)
	var videoTitle string
	if len(titleMatch) > 1 {
		videoTitle = string(titleMatch[1])
		videoTitle = html.UnescapeString(videoTitle)
	}

	splittedHTML := strings.Split(string(videoPageBody), `"captions":`)
	if len(splittedHTML) <= 1 {
		if strings.Contains(string(videoPageBody), `class="g-recaptcha"`) {
			return nil, "", &YoutubeTranscriptError{Message: "YouTube is receiving too many requests from this IP and now requires solving a captcha to continue"}
		}
		if !strings.Contains(string(videoPageBody), `"playabilityStatus":`) {
			return nil, "", &YoutubeTranscriptError{Message: fmt.Sprintf("The video is no longer available (%s)", videoId)}
		}
		return nil, "", &YoutubeTranscriptError{Message: fmt.Sprintf("Transcript is disabled on this video (%s)", videoId)}
	}

	var captions struct {
		PlayerCaptionsTracklistRenderer struct {
			CaptionTracks []struct {
				BaseURL      string `json:"baseUrl"`
				LanguageCode string `json:"languageCode"`
			} `json:"captionTracks"`
		} `json:"playerCaptionsTracklistRenderer"`
	}

	captionsData := splittedHTML[1][:strings.Index(splittedHTML[1], ",\"videoDetails")]
	err = json.Unmarshal([]byte(captionsData), &captions)
	if err != nil {
		return nil, "", &YoutubeTranscriptError{Message: fmt.Sprintf("Transcript is disabled on this video (%s)", videoId)}
	}

	if len(captions.PlayerCaptionsTracklistRenderer.CaptionTracks) == 0 {
		return nil, "", &YoutubeTranscriptError{Message: fmt.Sprintf("No transcripts are available for this video (%s)", videoId)}
	}

	transcriptURL := captions.PlayerCaptionsTracklistRenderer.CaptionTracks[0].BaseURL

	transcriptResponse, err := services.DefaultHttpClient().Get(transcriptURL)
	if err != nil {
		return nil, "", &YoutubeTranscriptError{Message: fmt.Sprintf("No transcripts are available for this video (%s)", videoId)}
	}
	defer transcriptResponse.Body.Close()

	transcriptBody, err := io.ReadAll(transcriptResponse.Body)
	if err != nil {
		return nil, "", err
	}

	re := regexp.MustCompile(RE_XML_TRANSCRIPT)
	matches := re.FindAllStringSubmatch(string(transcriptBody), -1)
	var results []TranscriptResponse
	for _, match := range matches {
		duration, _ := strconv.ParseFloat(match[2], 64)
		offset, _ := strconv.ParseFloat(match[1], 64)
		results = append(results, TranscriptResponse{
			Text:     match[3],
			Duration: duration,
			Offset:   offset,
			Lang:     captions.PlayerCaptionsTracklistRenderer.CaptionTracks[0].LanguageCode,
		})
	}

	return results, videoTitle, nil
}

// Helper functions
func retrieveVideoId(videoId string) (string, error) {
	if len(videoId) == 11 {
		return videoId, nil
	}
	re := regexp.MustCompile(RE_YOUTUBE)
	match := re.FindStringSubmatch(videoId)
	if match != nil {
		return match[1], nil
	}
	return "", &YoutubeTranscriptError{Message: "Impossible to retrieve Youtube video ID."}
}

func decodeHTML(text string) string {
	text = strings.ReplaceAll(text, "&amp;#39;", "'")
	text = html.UnescapeString(text)
	return text
}

func formatTimestamp(offset float64) string {
	duration := time.Duration(offset * float64(time.Second))
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute
	duration -= minutes * time.Minute
	seconds := duration / time.Second
	return fmt.Sprintf("[%02d:%02d:%02d] ", hours, minutes, seconds)
}
