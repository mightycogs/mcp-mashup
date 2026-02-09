package aggregator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mightycogs/mcp-mashup/pkg/config"
	"github.com/mightycogs/mcp-mashup/pkg/logger"
)

var Version = "dev"

type MCPClient interface {
	Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error)
	ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error)
	CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	Close() error
}

type MCPAggregator struct {
	clients map[string]MCPClient
	tools   map[string]toolMapping
	configs map[string]*config.ServerConfig
	mu      sync.RWMutex
}

type toolMapping struct {
	serverName    string
	originalName  string
	sanitizedName string
	tool          mcp.Tool
}

func sanitizeName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func NewMCPAggregator() *MCPAggregator {
	return &MCPAggregator{
		clients: make(map[string]MCPClient),
		tools:   make(map[string]toolMapping),
		configs: make(map[string]*config.ServerConfig),
	}
}

func (a *MCPAggregator) Initialize(ctx context.Context, cfg *config.Config) error {
	for _, serverCfg := range cfg.Servers {
		a.mu.Lock()
		a.configs[serverCfg.Name] = &serverCfg
		a.mu.Unlock()

		var envVars []string
		for key, value := range serverCfg.Env {
			envVars = append(envVars, key+"="+value)
		}

		logger.Debug("Initializing MCP server %s with command: %s %v", serverCfg.Name, serverCfg.Command, serverCfg.Args)
		logger.Debug("Environment variables: %v", envVars)

		mcpClient, err := client.NewStdioMCPClient(
			serverCfg.Command,
			envVars,
			serverCfg.Args...,
		)
		if err != nil {
			logger.Error("Failed to create client for server %s: %v", serverCfg.Name, err)
			return fmt.Errorf("failed to create client for server %s: %w", serverCfg.Name, err)
		}

		ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)

		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "mcp-mashup",
			Version: Version,
		}

		logger.Debug("Sending initialize request to %s...", serverCfg.Name)
		initResult, err := mcpClient.Initialize(ctxWithTimeout, initRequest)
		cancel()
		if err != nil {
			mcpClient.Close()
			logger.Error("Failed to initialize server %s: %v", serverCfg.Name, err)

			if ctxWithTimeout.Err() != nil || strings.Contains(err.Error(), "context") {
				logger.Error("Context error for server %s: %v", serverCfg.Name, err)
				logger.Error("Skipping server %s", serverCfg.Name)
				continue
			}

			logger.Error("Error initializing server %s: %v", serverCfg.Name, err)
			logger.Error("Continuing with other servers...")
			continue
		}
		logger.Info("Server %s initialized: %s %s", serverCfg.Name, initResult.ServerInfo.Name, initResult.ServerInfo.Version)

		a.mu.Lock()
		a.clients[serverCfg.Name] = mcpClient
		a.mu.Unlock()

		err = a.discoverTools(ctx, serverCfg.Name)
		if err != nil {
			logger.Error("Failed to discover tools for server %s: %v", serverCfg.Name, err)
			logger.Error("Continuing with other servers...")
			continue
		}
	}

	if len(a.clients) == 0 {
		return fmt.Errorf("no servers were successfully initialized")
	}

	return nil
}

func (a *MCPAggregator) discoverTools(ctx context.Context, serverName string) error {
	a.mu.RLock()
	mcpClient, exists := a.clients[serverName]
	serverConfig := a.configs[serverName]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client for server %s not found", serverName)
	}

	logger.Debug("Discovering tools for server %s...", serverName)
	toolsResp, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools for server %s: %w", serverName, err)
	}
	logger.Debug("Found %d tools for server %s", len(toolsResp.Tools), serverName)

	allowedTools := make(map[string]bool)
	if serverConfig != nil && serverConfig.Tools != nil {
		logger.Debug("Tool filtering enabled for server %s", serverName)
		for _, tool := range serverConfig.Tools.Allowed {
			normalizedName := sanitizeName(tool)
			logger.Debug("Adding allowed tool: %s (normalized: %s)", tool, normalizedName)
			allowedTools[normalizedName] = true
		}
		if len(serverConfig.Tools.Allowed) == 0 {
			logger.Debug("Empty allowed tools list for server %s, no tools will be exposed", serverName)
			return nil
		}
	} else {
		logger.Debug("No tool filtering configured for server %s", serverName)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sanitizedServerName := sanitizeName(serverName)
	for _, tool := range toolsResp.Tools {
		if len(allowedTools) > 0 {
			normalizedName := sanitizeName(tool.Name)
			if !allowedTools[normalizedName] {
				logger.Debug("Skipping tool %s (normalized: %s) as it's not in allowed list for server %s", tool.Name, normalizedName, serverName)
				continue
			}
			logger.Debug("Including allowed tool %s (normalized: %s) for server %s", tool.Name, normalizedName, serverName)
		}

		originalName := tool.Name
		sanitizedName := sanitizeName(originalName)
		prefixedName := fmt.Sprintf("%s_%s", sanitizedServerName, sanitizedName)

		logger.Debug("Registering tool: %s -> %s (sanitized from: %s)", originalName, prefixedName, tool.Name)

		a.tools[prefixedName] = toolMapping{
			serverName:    serverName,
			originalName:  originalName,
			sanitizedName: sanitizedName,
			tool:          tool,
		}
	}

	return nil
}

func (a *MCPAggregator) GetTools() []mcp.Tool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var allTools []mcp.Tool
	for prefixedName, mapping := range a.tools {
		tool := mapping.tool
		tool.Name = prefixedName

		if tool.Description != "" {
			tool.Description = fmt.Sprintf("[%s] %s", mapping.serverName, tool.Description)
		}

		ensureValidToolSchema(&tool)

		allTools = append(allTools, tool)
	}

	return allTools
}

func ensureValidToolSchema(tool *mcp.Tool) {
	if tool.InputSchema.Type == "" {
		tool.InputSchema.Type = "object"
	}

	if tool.InputSchema.Properties == nil {
		tool.InputSchema.Properties = make(map[string]interface{})
	}
}

func (a *MCPAggregator) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a.mu.RLock()
	prefixedName := request.Params.Name
	mapping, exists := a.tools[prefixedName]
	mcpClient, clientExists := a.clients[mapping.serverName]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", prefixedName)
	}

	if !clientExists {
		return nil, fmt.Errorf("client for server %s not found", mapping.serverName)
	}

	logger.Debug("Calling tool %s on server %s (mapped from %s)", mapping.originalName, mapping.serverName, prefixedName)

	newRequest := request
	newRequest.Params.Name = mapping.originalName

	return mcpClient.CallTool(ctx, newRequest)
}

func (a *MCPAggregator) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for name, mcpClient := range a.clients {
		mcpClient.Close()
		delete(a.clients, name)
	}
}
