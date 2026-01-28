package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	mcptools "github.com/metalpoch/local-synapse/internal/pkg/mcp_tools"
)

func main() {
	s := server.NewMCPServer(
		"local-synapse",
		"0.0.2",
		server.WithLogging(),
	)

	s.AddTool(mcptools.SystemStats())

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
