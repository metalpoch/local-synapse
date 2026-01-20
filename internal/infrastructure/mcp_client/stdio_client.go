package mcpclient

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/metalpoch/local-synapse/internal/domain"
)

type stdioClient struct {
	client client.MCPClient
}

// NewStdioClient creates a new MCP client that runs the given command
func NewStdioClient(command string, args ...string) (domain.MCPClient, error) {
	// NewStdioMCPClient(command string, env []string, args ...string)
	c, err := client.NewStdioMCPClient(command, nil, args...)
	if err != nil {
		return nil, err
	}
	return &stdioClient{
		client: c,
	}, nil
}

func (c *stdioClient) Initialize(ctx context.Context) error {
	// Client is already started by NewStdioMCPClient

	// Initialize the MCP session
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "local-synapse-api",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	_, err := c.client.Initialize(ctx, initRequest)
	if err != nil {
		return fmt.Errorf("failed to initialize mcp session: %w", err)
	}

	return nil
}

func (c *stdioClient) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	request := mcp.ListToolsRequest{}
	resp, err := c.client.ListTools(ctx, request)
	if err != nil {
		return nil, err
	}
	return resp.Tools, nil
}

func (c *stdioClient) CallTool(ctx context.Context, name string, args map[string]any) (*mcp.CallToolResult, error) {
	request := mcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = args
	
	resp, err := c.client.CallTool(ctx, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *stdioClient) Close() error {
	return c.client.Close()
}
