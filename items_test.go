package codex

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseItem_AgentMessage(t *testing.T) {
	raw := json.RawMessage(`{"id":"i1","type":"agent_message","text":"hi"}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	am, ok := item.(*AgentMessageItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if am.ID != "i1" || am.Text != "hi" {
		t.Errorf("got %+v", am)
	}
}

func TestParseItem_Reasoning(t *testing.T) {
	raw := json.RawMessage(`{"id":"i2","type":"reasoning","text":"thinking..."}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := item.(*ReasoningItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if r.Text != "thinking..." {
		t.Errorf("text = %q", r.Text)
	}
}

func TestParseItem_CommandExecution(t *testing.T) {
	raw := json.RawMessage(`{"id":"c1","type":"command_execution","command":"ls","aggregated_output":"a\nb","exit_code":0,"status":"completed"}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := item.(*CommandExecutionItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if c.Command != "ls" || c.AggregatedOutput != "a\nb" || c.Status != "completed" {
		t.Errorf("got %+v", c)
	}
	if c.ExitCode == nil || *c.ExitCode != 0 {
		t.Errorf("ExitCode = %v", c.ExitCode)
	}
}

func TestParseItem_FileChange(t *testing.T) {
	raw := json.RawMessage(`{"id":"f1","type":"file_change","status":"completed","changes":[{"path":"a.go","kind":"add"},{"path":"b.go","kind":"update"}]}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := item.(*FileChangeItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if len(f.Changes) != 2 || f.Changes[0].Path != "a.go" || f.Changes[1].Kind != "update" {
		t.Errorf("got %+v", f)
	}
}

func TestParseItem_McpToolCall(t *testing.T) {
	raw := json.RawMessage(`{"id":"m1","type":"mcp_tool_call","server":"s","tool":"t","arguments":{"k":1},"status":"completed"}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := item.(*McpToolCallItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if m.Server != "s" || m.Tool != "t" || m.Status != "completed" {
		t.Errorf("got %+v", m)
	}
	args, _ := m.Arguments.(map[string]any)
	if args == nil || args["k"].(float64) != 1 {
		t.Errorf("Arguments = %v", m.Arguments)
	}
}

func TestParseItem_WebSearch(t *testing.T) {
	raw := json.RawMessage(`{"id":"w1","type":"web_search","query":"go generics"}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	w, ok := item.(*WebSearchItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if w.Query != "go generics" {
		t.Errorf("Query = %q", w.Query)
	}
}

func TestParseItem_TodoList(t *testing.T) {
	raw := json.RawMessage(`{"id":"t1","type":"todo_list","items":[{"text":"a","completed":false},{"text":"b","completed":true}]}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	tl, ok := item.(*TodoListItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	want := []TodoItem{{Text: "a"}, {Text: "b", Completed: true}}
	if !reflect.DeepEqual(tl.Items, want) {
		t.Errorf("Items = %+v want %+v", tl.Items, want)
	}
}

func TestParseItem_Error(t *testing.T) {
	raw := json.RawMessage(`{"id":"e1","type":"error","message":"oops"}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	e, ok := item.(*ErrorItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if e.Message != "oops" {
		t.Errorf("Message = %q", e.Message)
	}
}

func TestParseItem_UnknownType(t *testing.T) {
	raw := json.RawMessage(`{"id":"x","type":"future_item","weird":42}`)
	item, err := parseItem(raw)
	if err != nil {
		t.Fatal(err)
	}
	u, ok := item.(*UnknownItem)
	if !ok {
		t.Fatalf("got %T", item)
	}
	if u.Type != "future_item" {
		t.Errorf("Type = %q", u.Type)
	}
	if string(u.Raw) == "" {
		t.Error("Raw empty")
	}
}

func TestParseItem_Malformed(t *testing.T) {
	_, err := parseItem(json.RawMessage(`not json`))
	if err == nil {
		t.Error("expected error")
	}
}
