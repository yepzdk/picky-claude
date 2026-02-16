package console

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/db"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// newMCPServer creates an MCP server with the memory tools registered.
func (s *Server) newMCPServer() *server.MCPServer {
	mcpSrv := server.NewMCPServer(
		config.DisplayName+" Memory",
		config.Version(),
	)

	mcpSrv.AddTool(searchTool(), s.handleMCPSearch)
	mcpSrv.AddTool(timelineTool(), s.handleMCPTimeline)
	mcpSrv.AddTool(getObservationsTool(), s.handleMCPGetObservations)
	mcpSrv.AddTool(saveMemoryTool(), s.handleMCPSaveMemory)

	return mcpSrv
}

func searchTool() mcp.Tool {
	return mcp.NewTool("search",
		mcp.WithDescription("Search observations by text query with optional filters"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 20)")),
		mcp.WithString("type", mcp.Description("Filter by type (bugfix, feature, refactor, discovery, decision, change)")),
		mcp.WithString("project", mcp.Description("Filter by project name")),
		mcp.WithString("dateStart", mcp.Description("Filter start date (YYYY-MM-DD)")),
		mcp.WithString("dateEnd", mcp.Description("Filter end date (YYYY-MM-DD)")),
	)
}

func timelineTool() mcp.Tool {
	return mcp.NewTool("timeline",
		mcp.WithDescription("Get chronological context around an observation ID"),
		mcp.WithNumber("anchor", mcp.Required(), mcp.Description("Observation ID to center on")),
		mcp.WithNumber("depth_before", mcp.Description("Observations before anchor (default 5)")),
		mcp.WithNumber("depth_after", mcp.Description("Observations after anchor (default 5)")),
	)
}

func getObservationsTool() mcp.Tool {
	return mcp.NewTool("get_observations",
		mcp.WithDescription("Fetch full details for specific observation IDs"),
		mcp.WithArray("ids", mcp.Required(), mcp.Description("Array of observation IDs")),
	)
}

func saveMemoryTool() mcp.Tool {
	return mcp.NewTool("save_memory",
		mcp.WithDescription("Save an observation to persistent memory"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Observation text")),
		mcp.WithString("title", mcp.Description("Short title")),
		mcp.WithString("project", mcp.Description("Project name")),
	)
}

func (s *Server) handleMCPSearch(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	query, _ := args["query"].(string)
	if query == "" {
		return mcpError("query parameter is required"), nil
	}

	filter := db.SearchFilter{
		Query: query,
		Limit: intArg(args, "limit", 20),
	}
	if v, ok := args["type"].(string); ok {
		filter.Type = v
	}
	if v, ok := args["project"].(string); ok {
		filter.Project = v
	}
	if v, ok := args["dateStart"].(string); ok {
		filter.DateStart = v
	}
	if v, ok := args["dateEnd"].(string); ok {
		filter.DateEnd = v
	}

	results, err := s.db.FilteredSearch(filter)
	if err != nil {
		return mcpError(fmt.Sprintf("search failed: %v", err)), nil
	}

	return mcpJSON(results)
}

func (s *Server) handleMCPTimeline(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	anchor := intArg(args, "anchor", 0)
	if anchor <= 0 {
		return mcpError("anchor parameter is required"), nil
	}

	before := intArg(args, "depth_before", 5)
	after := intArg(args, "depth_after", 5)

	results, err := s.db.TimelineAround(int64(anchor), before, after)
	if err != nil {
		return mcpError(fmt.Sprintf("timeline failed: %v", err)), nil
	}

	return mcpJSON(results)
}

func (s *Server) handleMCPGetObservations(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	rawIDs, ok := args["ids"].([]any)
	if !ok || len(rawIDs) == 0 {
		return mcpError("ids parameter is required (array of numbers)"), nil
	}

	ids := make([]int64, 0, len(rawIDs))
	for _, raw := range rawIDs {
		if v, ok := raw.(float64); ok {
			ids = append(ids, int64(v))
		}
	}

	results, err := s.db.GetObservations(ids)
	if err != nil {
		return mcpError(fmt.Sprintf("get_observations failed: %v", err)), nil
	}

	return mcpJSON(results)
}

func (s *Server) handleMCPSaveMemory(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	text, _ := args["text"].(string)
	if text == "" {
		return mcpError("text parameter is required"), nil
	}

	title, _ := args["title"].(string)
	project, _ := args["project"].(string)

	id, err := s.db.InsertObservation(&db.Observation{
		Type:    "discovery",
		Title:   title,
		Text:    text,
		Project: project,
	})
	if err != nil {
		return mcpError(fmt.Sprintf("save_memory failed: %v", err)), nil
	}

	return mcpJSON(map[string]int64{"id": id})
}

func mcpError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: msg},
		},
		IsError: true,
	}
}

func mcpJSON(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return mcpError(fmt.Sprintf("marshal error: %v", err)), nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: string(data)},
		},
	}, nil
}

func intArg(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}
