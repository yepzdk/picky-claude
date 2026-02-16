package console

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMCPToolsRegistered(t *testing.T) {
	srv := testServer(t)
	mcpSrv := srv.newMCPServer()

	tools := []string{"search", "timeline", "get_observations", "save_memory"}
	for _, name := range tools {
		tool := mcpSrv.GetTool(name)
		if tool == nil {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestMCPSearch(t *testing.T) {
	srv := testServer(t)

	// Insert test data
	doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1", "type": "discovery",
		"title": "auth bug", "text": "Found authentication issue",
		"project": "backend",
	})

	mcpSrv := srv.newMCPServer()
	tool := mcpSrv.GetTool("search")
	if tool == nil {
		t.Fatal("search tool not found")
	}

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "search",
			Arguments: map[string]any{"query": "authentication", "limit": float64(10)},
		},
	})
	if err != nil {
		t.Fatalf("search handler: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	tc, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatal("expected TextContent")
	}
	if tc.Text == "" {
		t.Error("expected non-empty text")
	}
}

func TestMCPGetObservations(t *testing.T) {
	srv := testServer(t)

	rr := doRequest(t, srv, "POST", "/api/observations", map[string]string{
		"session_id": "s1", "type": "discovery",
		"title": "test", "text": "test text",
	})
	var created map[string]any
	json.NewDecoder(rr.Body).Decode(&created)
	id := created["id"].(float64)

	mcpSrv := srv.newMCPServer()
	tool := mcpSrv.GetTool("get_observations")
	if tool == nil {
		t.Fatal("get_observations tool not found")
	}

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_observations",
			Arguments: map[string]any{"ids": []any{id}},
		},
	})
	if err != nil {
		t.Fatalf("get_observations handler: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestMCPTimeline(t *testing.T) {
	srv := testServer(t)

	for i := 0; i < 5; i++ {
		doRequest(t, srv, "POST", "/api/observations", map[string]string{
			"session_id": "s1", "type": "discovery",
			"title": "obs", "text": "text",
		})
	}

	mcpSrv := srv.newMCPServer()
	tool := mcpSrv.GetTool("timeline")
	if tool == nil {
		t.Fatal("timeline tool not found")
	}

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "timeline",
			Arguments: map[string]any{"anchor": float64(3)},
		},
	})
	if err != nil {
		t.Fatalf("timeline handler: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestMCPSaveMemory(t *testing.T) {
	srv := testServer(t)

	mcpSrv := srv.newMCPServer()
	tool := mcpSrv.GetTool("save_memory")
	if tool == nil {
		t.Fatal("save_memory tool not found")
	}

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "save_memory",
			Arguments: map[string]any{
				"text":  "Important discovery about auth",
				"title": "Auth Discovery",
			},
		},
	})
	if err != nil {
		t.Fatalf("save_memory handler: %v", err)
	}
	if result.IsError {
		t.Error("expected successful result")
	}
}
