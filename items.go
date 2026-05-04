package codex

import (
	"encoding/json"
	"fmt"
)

// ThreadItem is a sealed interface implemented by every item variant.
// Mirror of TS `ThreadItem` discriminated union (items.ts).
type ThreadItem interface{ threadItem() }

type AgentMessageItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

type ReasoningItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

// CommandExecutionStatus mirrors TS.
type CommandExecutionStatus string

const (
	CmdInProgress CommandExecutionStatus = "in_progress"
	CmdCompleted  CommandExecutionStatus = "completed"
	CmdFailed     CommandExecutionStatus = "failed"
)

type CommandExecutionItem struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Command          string                 `json:"command"`
	AggregatedOutput string                 `json:"aggregated_output"`
	ExitCode         *int                   `json:"exit_code,omitempty"`
	Status           CommandExecutionStatus `json:"status"`
}

type PatchChangeKind string

const (
	PatchAdd    PatchChangeKind = "add"
	PatchDelete PatchChangeKind = "delete"
	PatchUpdate PatchChangeKind = "update"
)

type FileUpdateChange struct {
	Path string          `json:"path"`
	Kind PatchChangeKind `json:"kind"`
}

type PatchApplyStatus string

const (
	PatchCompleted PatchApplyStatus = "completed"
	PatchFailed    PatchApplyStatus = "failed"
)

type FileChangeItem struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Changes []FileUpdateChange `json:"changes"`
	Status  PatchApplyStatus   `json:"status"`
}

type McpToolCallStatus string

const (
	McpInProgress McpToolCallStatus = "in_progress"
	McpCompleted  McpToolCallStatus = "completed"
	McpFailed     McpToolCallStatus = "failed"
)

type McpToolCallItem struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Server    string            `json:"server"`
	Tool      string            `json:"tool"`
	Arguments any               `json:"arguments"`
	Result    *McpToolCallResult `json:"result,omitempty"`
	Error     *McpToolCallError  `json:"error,omitempty"`
	Status    McpToolCallStatus  `json:"status"`
}

type McpToolCallResult struct {
	// Content is the MCP server's content blocks. Left as raw JSON because
	// modeling MCP content blocks fully would re-implement
	// @modelcontextprotocol/sdk; consumers can decode further if needed.
	Content           json.RawMessage `json:"content"`
	StructuredContent any             `json:"structured_content"`
}

type McpToolCallError struct {
	Message string `json:"message"`
}

type WebSearchItem struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Query string `json:"query"`
}

type TodoItem struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type TodoListItem struct {
	ID    string     `json:"id"`
	Type  string     `json:"type"`
	Items []TodoItem `json:"items"`
}

type ErrorItem struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// UnknownItem is emitted when an item.* event carries a "type" the SDK
// does not recognize. Forward-compat.
type UnknownItem struct {
	Type string
	Raw  json.RawMessage
}

func (*AgentMessageItem) threadItem()     {}
func (*ReasoningItem) threadItem()        {}
func (*CommandExecutionItem) threadItem() {}
func (*FileChangeItem) threadItem()       {}
func (*McpToolCallItem) threadItem()      {}
func (*WebSearchItem) threadItem()        {}
func (*TodoListItem) threadItem()         {}
func (*ErrorItem) threadItem()            {}
func (*UnknownItem) threadItem()          {}

// parseItem decodes one item JSON object. Returns *UnknownItem for unknown
// types (forward-compat) and an error for malformed JSON or schema errors
// inside a known type.
func parseItem(raw json.RawMessage) (ThreadItem, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("parseItem: empty raw")
	}
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, fmt.Errorf("parseItem head: %w", err)
	}
	switch head.Type {
	case "agent_message":
		var v AgentMessageItem
		return &v, json.Unmarshal(raw, &v)
	case "reasoning":
		var v ReasoningItem
		return &v, json.Unmarshal(raw, &v)
	case "command_execution":
		var v CommandExecutionItem
		return &v, json.Unmarshal(raw, &v)
	case "file_change":
		var v FileChangeItem
		return &v, json.Unmarshal(raw, &v)
	case "mcp_tool_call":
		var v McpToolCallItem
		return &v, json.Unmarshal(raw, &v)
	case "web_search":
		var v WebSearchItem
		return &v, json.Unmarshal(raw, &v)
	case "todo_list":
		var v TodoListItem
		return &v, json.Unmarshal(raw, &v)
	case "error":
		var v ErrorItem
		return &v, json.Unmarshal(raw, &v)
	default:
		return &UnknownItem{Type: head.Type, Raw: append([]byte(nil), raw...)}, nil
	}
}
