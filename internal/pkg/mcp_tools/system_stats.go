package mcptools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	systemmetrics "github.com/metalpoch/local-synapse/internal/infrastructure/system_metrics"
)

func SystemStats() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("system-stats", mcp.WithDescription("Get system metrics (CPU, RAM, Disk, Network)")),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			metrics, err := systemmetrics.GetSystemMetrics()
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get system metrics: %v", err)), nil
			}
			response := fmt.Sprintf(
				"CPU: %.2f%%\nRAM: %.2f%% (Used: %dGB / Total: %dGB)\nDisk: %.2f%% (Used: %dGB / Total: %dGB)\nNetwork: Sent %d bytes, Recv %d bytes",
				metrics.CPUPercent[0],
				metrics.RAM.Usage, metrics.RAM.Used, metrics.RAM.Total,
				metrics.Disk.Usage, metrics.Disk.Used, metrics.Disk.Total,
				metrics.Network.BytesSent, metrics.Network.BytesRecv,
			)
			return mcp.NewToolResultText(response), nil
		}

}
