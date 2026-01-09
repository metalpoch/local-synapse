package mcp

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	systemmetrics "github.com/metalpoch/local-synapse/internal/infrastructure/system_metrics"
)

func NewServer() *server.MCPServer {
	// Initialize the MCP server
	s := server.NewMCPServer(
		"local-synapse",
		"0.0.1",
		server.WithLogging(),
	)

	// Register the system-stats tool
	s.AddTool(mcp.NewTool("system-stats",
		mcp.WithDescription("Get system metrics (CPU, RAM, Disk, Network)"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		metrics, err := systemmetrics.GetSystemMetrics()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get system metrics: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("CPU: %.2f%%\nRAM: %.2f%% (Used: %dGB / Total: %dGB)\nDisk: %.2f%% (Used: %dGB / Total: %dGB)\nNetwork: Sent %d bytes, Recv %d bytes",
			metrics.CPUPercent[0],
			metrics.RAM.Usage, metrics.RAM.Used, metrics.RAM.Total,
			metrics.Disk.Usage, metrics.Disk.Used, metrics.Disk.Total,
			metrics.Network.BytesSent, metrics.Network.BytesRecv,
		)), nil
	})

	return s
}

// ServeStdio starts the MCP server over standard input/output
func ServeStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}

// Run starts the server using Stdio
func Run() {
	s := NewServer()
	if err := ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
