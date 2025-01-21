package tools

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ThoughtData struct {
	Thought           string  `json:"thought"`
	ThoughtNumber     int     `json:"thoughtNumber"`
	TotalThoughts     int     `json:"totalThoughts"`
	IsRevision        *bool   `json:"isRevision,omitempty"`
	RevisesThought    *int    `json:"revisesThought,omitempty"`
	BranchFromThought *int    `json:"branchFromThought,omitempty"`
	BranchID          *string `json:"branchId,omitempty"`
	NeedsMoreThoughts *bool   `json:"needsMoreThoughts,omitempty"`
	NextThoughtNeeded bool    `json:"nextThoughtNeeded"`
}

type SequentialThinkingServer struct {
	thoughtHistory []ThoughtData
	branches      map[string][]ThoughtData
}

func NewSequentialThinkingServer() *SequentialThinkingServer {
	return &SequentialThinkingServer{
		thoughtHistory: make([]ThoughtData, 0),
		branches:      make(map[string][]ThoughtData),
	}
}

func (s *SequentialThinkingServer) validateThoughtData(input map[string]interface{}) (*ThoughtData, error) {
	thought, ok := input["thought"].(string)
	if !ok || thought == "" {
		return nil, fmt.Errorf("invalid thought: must be a string")
	}

	thoughtNumber, ok := input["thoughtNumber"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid thoughtNumber: must be a number")
	}

	totalThoughts, ok := input["totalThoughts"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid totalThoughts: must be a number")
	}

	nextThoughtNeeded, ok := input["nextThoughtNeeded"].(bool)
	if !ok {
		return nil, fmt.Errorf("invalid nextThoughtNeeded: must be a boolean")
	}

	data := &ThoughtData{
		Thought:           thought,
		ThoughtNumber:     int(thoughtNumber),
		TotalThoughts:     int(totalThoughts),
		NextThoughtNeeded: nextThoughtNeeded,
	}

	// Optional fields
	if isRevision, ok := input["isRevision"].(bool); ok {
		data.IsRevision = &isRevision
	}
	if revisesThought, ok := input["revisesThought"].(float64); ok {
		rt := int(revisesThought)
		data.RevisesThought = &rt
	}
	if branchFromThought, ok := input["branchFromThought"].(float64); ok {
		bft := int(branchFromThought)
		data.BranchFromThought = &bft
	}
	if branchID, ok := input["branchId"].(string); ok {
		data.BranchID = &branchID
	}
	if needsMoreThoughts, ok := input["needsMoreThoughts"].(bool); ok {
		data.NeedsMoreThoughts = &needsMoreThoughts
	}

	return data, nil
}

func (s *SequentialThinkingServer) processThought(input map[string]interface{}) (*mcp.CallToolResult, error) {
	thoughtData, err := s.validateThoughtData(input)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if thoughtData.ThoughtNumber > thoughtData.TotalThoughts {
		thoughtData.TotalThoughts = thoughtData.ThoughtNumber
	}

	s.thoughtHistory = append(s.thoughtHistory, *thoughtData)

	if thoughtData.BranchFromThought != nil && thoughtData.BranchID != nil {
		if _, exists := s.branches[*thoughtData.BranchID]; !exists {
			s.branches[*thoughtData.BranchID] = make([]ThoughtData, 0)
		}
		s.branches[*thoughtData.BranchID] = append(s.branches[*thoughtData.BranchID], *thoughtData)
	}

	branchKeys := make([]string, 0, len(s.branches))
	for k := range s.branches {
		branchKeys = append(branchKeys, k)
	}

	response := map[string]interface{}{
		"thoughtNumber":       thoughtData.ThoughtNumber,
		"totalThoughts":      thoughtData.TotalThoughts,
		"nextThoughtNeeded":  thoughtData.NextThoughtNeeded,
		"branches":           branchKeys,
		"thoughtHistoryLength": len(s.thoughtHistory),
	}

	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

func RegisterSequentialThinkingTool(s *server.MCPServer) {
	thinkingServer := NewSequentialThinkingServer()

	sequentialThinkingTool := mcp.NewTool("sequentialthinking",
		mcp.WithDescription(`A detailed tool for dynamic and reflective problem-solving through thoughts.
This tool helps analyze problems through a flexible thinking process that can adapt and evolve.
Each thought can build on, question, or revise previous insights as understanding deepens.

When to use this tool:
- Breaking down complex problems into steps
- Planning and design with room for revision
- Analysis that might need course correction
- Problems where the full scope might not be clear initially
- Problems that require a multi-step solution
- Tasks that need to maintain context over multiple steps
- Situations where irrelevant information needs to be filtered out

Key features:
- You can adjust total_thoughts up or down as you progress
- You can question or revise previous thoughts
- You can add more thoughts even after reaching what seemed like the end
- You can express uncertainty and explore alternative approaches
- Not every thought needs to build linearly - you can branch or backtrack
- Generates a solution hypothesis
- Verifies the hypothesis based on the Chain of Thought steps
- Repeats the process until satisfied
- Provides a correct answer

Parameters explained:
- thought: Your current thinking step, which can include:
* Regular analytical steps
* Revisions of previous thoughts
* Questions about previous decisions
* Realizations about needing more analysis
* Changes in approach
* Hypothesis generation
* Hypothesis verification
- next_thought_needed: True if you need more thinking, even if at what seemed like the end
- thought_number: Current number in sequence (can go beyond initial total if needed)
- total_thoughts: Current estimate of thoughts needed (can be adjusted up/down)
- is_revision: A boolean indicating if this thought revises previous thinking
- revises_thought: If is_revision is true, which thought number is being reconsidered
- branch_from_thought: If branching, which thought number is the branching point
- branch_id: Identifier for the current branch (if any)
- needs_more_thoughts: If reaching end but realizing more thoughts needed

You should:
1. Start with an initial estimate of needed thoughts, but be ready to adjust
2. Feel free to question or revise previous thoughts
3. Don't hesitate to add more thoughts if needed, even at the "end"
4. Express uncertainty when present
5. Mark thoughts that revise previous thinking or branch into new paths
6. Ignore information that is irrelevant to the current step
7. Generate a solution hypothesis when appropriate
8. Verify the hypothesis based on the Chain of Thought steps
9. Repeat the process until satisfied with the solution
10. Provide a single, ideally correct answer as the final output
11. Only set next_thought_needed to false when truly done and a satisfactory answer is reached`),
		mcp.WithString("thought", mcp.Required(), mcp.Description("Your current thinking step")),
		mcp.WithBoolean("nextThoughtNeeded", mcp.Required(), mcp.Description("Whether another thought step is needed")),
		mcp.WithNumber("thoughtNumber", mcp.Required(), mcp.Description("Current thought number")),
		mcp.WithNumber("totalThoughts", mcp.Required(), mcp.Description("Estimated total thoughts needed")),
		mcp.WithBoolean("isRevision", mcp.Description("Whether this revises previous thinking")),
		mcp.WithNumber("revisesThought", mcp.Description("Which thought is being reconsidered")),
		mcp.WithNumber("branchFromThought", mcp.Description("Branching point thought number")),
		mcp.WithString("branchId", mcp.Description("Branch identifier")),
		mcp.WithBoolean("needsMoreThoughts", mcp.Description("If more thoughts are needed")),
	)

	s.AddTool(sequentialThinkingTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		return thinkingServer.processThought(arguments)
	})
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func repeatStr(s string, n int) string {
	result := make([]byte, n*len(s))
	for i := 0; i < n; i++ {
		copy(result[i*len(s):], s)
	}
	return string(result)
}
