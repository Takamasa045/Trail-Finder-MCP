package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Takamasa045/Trail-Finder-MCP/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPToolDiscoveryAndCalls(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := NewServer().Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 6 {
		t.Fatalf("tools=%d", len(tools.Tools))
	}
	for _, tool := range tools.Tools {
		if tool.OutputSchema == nil {
			t.Errorf("%s has no output schema", tool.Name)
		}
		raw, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatal(err)
		}
		var schema struct {
			Properties map[string]struct{ Description string }
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatal(err)
		}
		for name, prop := range schema.Properties {
			if prop.Description == "" {
				t.Errorf("%s.%s has no description", tool.Name, name)
			}
		}
	}

	// Required coordinates must be rejected by the schema, before any upstream request.
	if result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "elevation", Arguments: map[string]any{}}); err == nil && !result.IsError {
		t.Fatal("missing coordinates accepted")
	}
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "forecast", Arguments: map[string]any{"lat": 91, "lon": 139}})
	if err != nil || !result.IsError {
		t.Fatalf("validation: result=%+v err=%v", result, err)
	}

	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/broken" {
			_, _ = w.Write([]byte(`{"remark":"runtime error: Query timed out","elements":[]}`))
			return
		}
		_, _ = w.Write([]byte(`[{"display_name":"高尾山口駅, 八王子市","lat":"35.6319","lon":"139.2694"}]`))
	}))
	defer up.Close()
	t.Setenv("NOMINATIM_URL", up.URL)
	t.Setenv("NOMINATIM_MIN_INTERVAL", "0")
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "geocode", Arguments: map[string]any{"query": "高尾山口駅"}})
	if err != nil || result.IsError {
		t.Fatalf("success: result=%+v err=%v", result, err)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var resp models.GeocodeResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Location.Lat != 35.6319 || len(result.Content) == 0 {
		t.Fatalf("missing structured/text result: %+v", result)
	}
	t.Setenv("TRAILFINDER_OVERPASS_URL", up.URL+"/broken")
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "trailheads", Arguments: map[string]any{"lat": 35, "lon": 139}})
	if err != nil || !result.IsError {
		t.Fatalf("upstream failure: result=%+v err=%v", result, err)
	}
	if text, ok := result.Content[0].(*mcp.TextContent); !ok || !strings.Contains(text.Text, "timed out") {
		t.Fatalf("error message missing: %+v", result.Content)
	}
}
