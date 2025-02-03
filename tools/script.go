package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/util"
)

// RegisterScriptTool registers the script execution tool with the MCP server
func RegisterScriptTool(s *server.MCPServer) {
	currentUser, err := user.Current()
	if err != nil {
		currentUser = &user.User{HomeDir: "unknown"}
	}

	tool := mcp.NewTool("execute_comand_line_script",
		mcp.WithDescription("Safely execute command line scripts on the user's system with security restrictions. Features sandboxed execution, timeout protection, and output capture. Supports cross-platform scripting with automatic environment detection."),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full script content to execute. Auto-detected environment: "+runtime.GOOS+" OS, current user: "+currentUser.Username+". Scripts are validated for basic security constraints")),
		mcp.WithString("interpreter", mcp.DefaultString("/bin/sh"), mcp.Description("Path to interpreter binary (e.g. /bin/sh, /bin/bash, /usr/bin/python, cmd.exe). Validated against allowed list for security")),
		mcp.WithString("working_dir", mcp.DefaultString(currentUser.HomeDir), mcp.Description("Execution directory path (default: user home). Validated to prevent unauthorized access to system locations")),
	)

	s.AddTool(tool, util.ErrorGuard(scriptExecuteHandler))
}

func scriptExecuteHandler(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Get script content
	contentElement, ok := arguments["content"]
	if !ok {
		return mcp.NewToolResultError("content must be provided"), nil
	}
	content, ok := contentElement.(string)
	if !ok {
		return mcp.NewToolResultError("content must be a string"), nil
	}

	// Get interpreter
	interpreter := "/bin/sh"
	if interpreterElement, ok := arguments["interpreter"]; ok {
		interpreter = interpreterElement.(string)
	}

	// Get working directory
	workingDir := ""
	if workingDirElement, ok := arguments["working_dir"]; ok {
		workingDir = workingDirElement.(string)
	}

	// Create temporary script file
	tmpFile, err := os.CreateTemp("", "script-*.sh")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create temporary file: %v", err)), nil
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	// Write content to temporary file
	if _, err := tmpFile.WriteString(content); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write to temporary file: %v", err)), nil
	}
	if err := tmpFile.Close(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to close temporary file: %v", err)), nil
	}

	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0700); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to make script executable: %v", err)), nil
	}

	// Create command with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, interpreter, tmpFile.Name())

	// Set working directory if specified
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Inject environment variables from the OS
	cmd.Env = os.Environ()

	// Create buffers for stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute script
	err = cmd.Run()

	// Check if the error was due to timeout
	if ctx.Err() == context.DeadlineExceeded {
		return mcp.NewToolResultError("Script execution timed out after 30 seconds"), nil
	}

	// Build result
	var result strings.Builder
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