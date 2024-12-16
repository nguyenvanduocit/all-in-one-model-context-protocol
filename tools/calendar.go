package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func RegisterCalendarTools(s *server.MCPServer) {
	// Create event tool
	createEventTool := mcp.NewTool("calendar_create_event",
		mcp.WithDescription("Create a new event in Google Calendar"),
		mcp.WithString("summary", mcp.Required(), mcp.Description("Title of the event")),
		mcp.WithString("description", mcp.Description("Description of the event")),
		mcp.WithString("start_time", mcp.Required(), mcp.Description("Start time of the event in RFC3339 format (e.g., 2023-12-25T09:00:00Z)")),
		mcp.WithString("end_time", mcp.Required(), mcp.Description("End time of the event in RFC3339 format")),
		mcp.WithString("attendees", mcp.Description("Comma-separated list of attendee email addresses")),
	)
	s.AddTool(createEventTool, util.ErrorGuard(calendarCreateEventHandler))

	// List events tool
	listEventsTool := mcp.NewTool("calendar_list_events",
		mcp.WithDescription("List upcoming events in Google Calendar"),
		mcp.WithString("time_min", mcp.Description("Start time for the search in RFC3339 format (default: now)")),
		mcp.WithString("time_max", mcp.Description("End time for the search in RFC3339 format (default: 1 week from now)")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of events to return (default: 10)")),
	)
	s.AddTool(listEventsTool, util.ErrorGuard(calendarListEventsHandler))

	// Update event tool
	updateEventTool := mcp.NewTool("calendar_update_event",
		mcp.WithDescription("Update an existing event in Google Calendar"),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("ID of the event to update")),
		mcp.WithString("summary", mcp.Description("New title of the event")),
		mcp.WithString("description", mcp.Description("New description of the event")),
		mcp.WithString("start_time", mcp.Description("New start time of the event in RFC3339 format")),
		mcp.WithString("end_time", mcp.Description("New end time of the event in RFC3339 format")),
		mcp.WithString("attendees", mcp.Description("Comma-separated list of new attendee email addresses")),
	)
	s.AddTool(updateEventTool, util.ErrorGuard(calendarUpdateEventHandler))

	// Respond to event tool
	respondToEventTool := mcp.NewTool("calendar_respond_to_event",
		mcp.WithDescription("Respond to an event invitation in Google Calendar"),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("ID of the event to respond to")),
		mcp.WithString("response", mcp.Required(), mcp.Description("Your response (accepted, declined, or tentative)")),
	)
	s.AddTool(respondToEventTool, util.ErrorGuard(calendarRespondToEventHandler))
}

var calendarService = sync.OnceValue[*calendar.Service](func() *calendar.Service {
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

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		panic(fmt.Sprintf("failed to create Calendar service: %v", err))
	}

	return srv
})

func calendarCreateEventHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	summary, _ := arguments["summary"].(string)
	description, _ := arguments["description"].(string)
	startTimeStr, _ := arguments["start_time"].(string)
	endTimeStr, _ := arguments["end_time"].(string)
	attendeesStr, _ := arguments["attendees"].(string)

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid start_time format"), nil
	}
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return mcp.NewToolResultError("Invalid end_time format"), nil
	}

	var attendees []*calendar.EventAttendee
	if attendeesStr != "" {
		for _, email := range strings.Split(attendeesStr, ",") {
			attendees = append(attendees, &calendar.EventAttendee{Email: email})
		}
	}

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		},
		Attendees: attendees,
	}

	createdEvent, err := calendarService().Events.Insert("primary", event).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create event: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created event with ID: %s", createdEvent.Id)), nil
}

func calendarListEventsHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	timeMinStr, ok := arguments["time_min"].(string)
	if !ok || timeMinStr == "" {
		timeMinStr = time.Now().Format(time.RFC3339)
	}

	timeMaxStr, ok := arguments["time_max"].(string)
	if !ok || timeMaxStr == "" {
		timeMaxStr = time.Now().AddDate(0, 0, 7).Format(time.RFC3339) // 1 week from now
	}

	maxResults, ok := arguments["max_results"].(float64)
	if !ok {
		maxResults = 10
	}

	events, err := calendarService().Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMinStr).
		TimeMax(timeMaxStr).
		MaxResults(int64(maxResults)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list events: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d upcoming events:\n\n", len(events.Items)))

	for _, item := range events.Items {
		start, _ := time.Parse(time.RFC3339, item.Start.DateTime)
		end, _ := time.Parse(time.RFC3339, item.End.DateTime)

		result.WriteString(fmt.Sprintf("Event: %s\n", item.Summary))
		result.WriteString(fmt.Sprintf("Start: %s\n", start.Format("2006-01-02 15:04")))
		result.WriteString(fmt.Sprintf("End: %s\n", end.Format("2006-01-02 15:04")))
		if item.Description != "" {
			result.WriteString(fmt.Sprintf("Description: %s\n", item.Description))
		}
		result.WriteString("-------------------\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func calendarUpdateEventHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	eventID, _ := arguments["event_id"].(string)
	summary, _ := arguments["summary"].(string)
	description, _ := arguments["description"].(string)
	startTimeStr, _ := arguments["start_time"].(string)
	endTimeStr, _ := arguments["end_time"].(string)
	attendeesStr, _ := arguments["attendees"].(string)

	event, err := calendarService().Events.Get("primary", eventID).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get event: %v", err)), nil
	}

	if summary != "" {
		event.Summary = summary
	}
	if description != "" {
		event.Description = description
	}
	if startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return mcp.NewToolResultError("Invalid start_time format"), nil
		}
		event.Start.DateTime = startTime.Format(time.RFC3339)
	}
	if endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return mcp.NewToolResultError("Invalid end_time format"), nil
		}
		event.End.DateTime = endTime.Format(time.RFC3339)
	}
	if attendeesStr != "" {
		var attendees []*calendar.EventAttendee
		for _, email := range strings.Split(attendeesStr, ",") {
			attendees = append(attendees, &calendar.EventAttendee{Email: email})
		}
		event.Attendees = attendees
	}

	updatedEvent, err := calendarService().Events.Update("primary", eventID, event).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update event: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully updated event with ID: %s", updatedEvent.Id)), nil
}

func calendarRespondToEventHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	eventID, _ := arguments["event_id"].(string)
	response, _ := arguments["response"].(string)

	event, err := calendarService().Events.Get("primary", eventID).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get event: %v", err)), nil
	}

	for _, attendee := range event.Attendees {
		if attendee.Self {
			attendee.ResponseStatus = response
			break
		}
	}

	_, err = calendarService().Events.Update("primary", eventID, event).Do()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update event response: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully responded '%s' to event with ID: %s", response, eventID)), nil
}