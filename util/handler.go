package util

import (
	"fmt"
	"runtime"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleError is a wrapper function that wraps the handler function with error handling
// Deprecated: Use ErrorGuard instead
func HandleError(handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return ErrorGuard(handler)
}

func ErrorGuard(handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(arguments map[string]interface{}) (result *mcp.CallToolResult, err error) {
		defer func() {
			if r := recover(); r != nil {
				// Get stack trace
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, true)
				stackTrace := string(buf[:n])
				
				result = mcp.NewToolResultError(fmt.Sprintf("Panic: %v\nStack trace:\n%s", r, stackTrace))
			}
		}()
		result, err = handler(arguments)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
		}
		return result, nil
	}
}
