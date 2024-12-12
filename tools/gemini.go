package tools

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/api/option"
)

func RegisterExpertTool(s *server.MCPServer) {
	tool := mcp.NewTool("ask_expert_gemini",
		mcp.WithDescription("Ask a question to an expert using Gemini"),
		mcp.WithString("question", mcp.Required(), mcp.Description("The question to ask")),
		// context
		mcp.WithString("context", mcp.Required(), mcp.Description("Context of the question")),
		mcp.WithString("session_id", mcp.Required(), mcp.Description("The session id to use, if not provided a new session will be created")),
	)

	s.AddTool(tool, commandLineExpertHandler)
}

var genAiClient = sync.OnceValue[*genai.Client](func() *genai.Client {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		panic("GOOGLE_AI_API_KEY environment variable must be set")
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		panic(fmt.Sprintf("failed to create Gemini client: %s", err))
	}

	return client
})

// sessionMap stores the chat sessions, key is the session id
var sessionMap = sync.Map{}

func commandLineExpertHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	question, ok := arguments["question"].(string)
	if !ok {
		return mcp.NewToolResultError("question must be a string"), nil
	}

	// get session id from arguments
	sessionId, ok := arguments["session_id"].(string)
	if !ok {
		sessionId = ""
	}

	model := genAiClient().GenerativeModel("gemini-2.0-flash-exp")
	model.SetTemperature(1)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "text/plain"

	var session *genai.ChatSession
	if sessionId == "" {
		sessionId = generateSessionId()
		session = model.StartChat()
		session.History = []*genai.Content{}
		sessionMap.Store(sessionId, session)
	} else {
		// load session from sessionMap
		if sess, ok := sessionMap.Load(sessionId); ok {
			session = sess.(*genai.ChatSession)
		} else {
			session = model.StartChat()
			session.History = []*genai.Content{}
			sessionMap.Store(sessionId, session)
		}
	}

	questionContext, ok := arguments["context"].(string)
	if !ok {
		questionContext = ""
	}

	question = "Context: " + questionContext + "\n\n" + "Question: " + question

	resp, err := session.SendMessage(context.Background(), genai.Text(question))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate content: %s", err)), nil
	}

	if len(resp.Candidates) == 0 {
		return mcp.NewToolResultError("no response from Gemini"), nil
	}

	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}



	return mcp.NewToolResultText("Session ID: " + sessionId + "\n\n" + text), nil
}

func generateSessionId() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to generate random session id: %s", err))
	}
	return hex.EncodeToString(b)
}
