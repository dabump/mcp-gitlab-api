package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/dabump/mcp-gitlab-api/internal/config"
	"github.com/dabump/mcp-gitlab-api/internal/gitlabapi"
	"github.com/dabump/mcp-gitlab-api/internal/tools"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to YAML configuration file")
	allowMissingConfig := flag.Bool("allow-missing-config", false, "Allow config file to be absent when env vars provide settings")
	flag.Parse()

	path := *configPath
	if *allowMissingConfig {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			path = ""
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	client, err := gitlabapi.New(cfg.GitLab)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GitLab client error: %v\n", err)
		os.Exit(1)
	}

	mcpServer := server.NewMCPServer("mcp-gitlab-api", "0.1.0")
	tools.RegisterAll(mcpServer, client)

	if err := serve(mcpServer, cfg.Server); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func serve(mcpServer *server.MCPServer, cfg config.ServerConfig) error {
	if cfg.Transport == "http" {
		addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
		httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath(cfg.Endpoint))
		return httpServer.Start(addr)
	}

	return server.ServeStdio(mcpServer)
}
