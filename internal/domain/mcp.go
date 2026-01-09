package domain

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// MCPClient defines the interface for an MCP client
type MCPClient interface {
	// Initialize starts the client and performs the handshake
	Initialize(ctx context.Context) error
	// ListTools returns the list of available tools
	ListTools(ctx context.Context) ([]mcp.Tool, error)
	// CallTool executes a tool and returns the result
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error)
	// Close terminates the client connection
	Close() error
}
