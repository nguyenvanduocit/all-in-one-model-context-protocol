package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
)

// RegisterCLITool registers the CLI tool with the MCP server
func RegisterCLITool(s *server.MCPServer) {

	currentUser, err := user.Current()
	if err != nil {
		currentUser = &user.User{HomeDir: "unknown"}
	}

	tool := mcp.NewTool("cli_execute",
		mcp.WithDescription("Execute a single command line operation on user machine (not a persistent session) on " + runtime.GOOS),
		mcp.WithString("command", mcp.Required(), mcp.Description("Command to execute")),
		mcp.WithString("args", mcp.DefaultString(""), mcp.Description("Command arguments (space-separated)")),
		mcp.WithString("working_dir", mcp.DefaultString(currentUser.HomeDir), mcp.Description("Working directory for command execution")),
	)

	s.AddTool(tool, util.ErrorGuard(cliExecuteHandler))
}

func cliExecuteHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {

	commandElement, ok := arguments["command"]
	if !ok {
		return mcp.NewToolResultError("command must be a string"), nil
	}
	command, ok := commandElement.(string)

	if !ok {
		return mcp.NewToolResultError("command must be a string"), nil
	}

	var argsStr string
	argsElement, ok := arguments["args"]
	if !ok {
		argsStr = ""
	} else {
		argsStr = argsElement.(string)
	}

	var workingDir string
	workingDirElement, ok := arguments["working_dir"]
	if !ok {
		workingDir = ""
	} else {
		workingDir = workingDirElement.(string)
	}

	// Split args string into slice, handling quoted arguments
	var args []string
	if argsStr != "" {
		args = parseCommandArgs(argsStr)
	}

	// Create command
	cmd := exec.Command(command, args...)

	// Set working directory if specified
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Create buffers for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	// Build result
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Single Command Execution:\n"))
	result.WriteString(fmt.Sprintf("Command: %s %s\n", command, argsStr))
	if workingDir != "" {
		result.WriteString(fmt.Sprintf("Working Directory: %s\n", workingDir))
	}
	result.WriteString("\n")

	if stdout.Len() > 0 {
		result.WriteString("Output:\n")
		result.WriteString(stdout.String())
		result.WriteString("\n")
	}

	if stderr.Len() > 0 {
		result.WriteString("Errors:\n")
		result.WriteString(stderr.String())
		result.WriteString("\n")
	}

	if err != nil {
		result.WriteString(fmt.Sprintf("\nExecution error: %v", err))
	}

	return mcp.NewToolResultText(result.String()), nil
}

// parseCommandArgs splits a command string into arguments, respecting quoted strings
func parseCommandArgs(argsStr string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, char := range argsStr {
		switch {
		case char == '"' || char == '\'':
			if inQuotes && char == quoteChar {
				// End quote
				inQuotes = false
				quoteChar = rune(0)
			} else if !inQuotes {
				// Start quote
				inQuotes = true
				quoteChar = char
			} else {
				// Quote character inside another type of quote
				currentArg.WriteRune(char)
			}
		case char == ' ' && !inQuotes:
			// Space outside quotes - split argument
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
		default:
			currentArg.WriteRune(char)
		}
	}

	// Add final argument if exists
	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}
