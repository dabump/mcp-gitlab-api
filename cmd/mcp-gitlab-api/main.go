package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

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
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}

		mcpHandler := server.NewStreamableHTTPServer(mcpServer)
		mux := http.NewServeMux()
		mux.Handle(normalizeHTTPPath(cfg.Endpoint), mcpHandler)

		_, port, err := net.SplitHostPort(listener.Addr().String())
		if err != nil {
			port = fmt.Sprintf("%d", cfg.Port)
		}
		fmt.Fprintf(os.Stdout, "MCP server started on port %s\n", port)

		httpServer := &http.Server{Handler: mux}
		return httpServer.Serve(listener)
	}

	return server.ServeStdio(mcpServer)
}

func normalizeHTTPPath(path string) string {
	return "/" + strings.Trim(path, "/")
}
