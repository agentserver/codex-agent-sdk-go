package codex

import (
	"encoding/json"
)

// ThreadEvent is a sealed interface implemented by every event variant.
// Mirror of TS `ThreadEvent` discriminated union (events.ts).
type ThreadEvent interface{ threadEvent() }

// Usage mirrors TS `Usage`.
type Usage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
}

// ThreadError mirrors TS `ThreadError`.
type ThreadError struct {
	Message string `json:"message"`
}

type ThreadStartedEvent struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id"`
}

type TurnStartedEvent struct {
	Type string `json:"type"`
}

type TurnCompletedEvent struct {
	Type  string `json:"type"`
	Usage Usage  `json:"usage"`
}

type TurnFailedEvent struct {
	Type  string      `json:"type"`
	Error ThreadError `json:"error"`
}

type ItemStartedEvent struct {
	Type string     `json:"type"`
	Item ThreadItem `json:"-"`
	raw  json.RawMessage
}

type ItemUpdatedEvent struct {
	Type string     `json:"type"`
	Item ThreadItem `json:"-"`
	raw  json.RawMessage
}

type ItemCompletedEvent struct {
	Type string     `json:"type"`
	Item ThreadItem `json:"-"`
	raw  json.RawMessage
}

type ThreadErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// UnknownEvent is emitted when a JSONL line has a "type" field whose value
// the SDK does not recognize. Forward-compat: codex CLI may add new event
// types ahead of the SDK.
type UnknownEvent struct {
	Type string
	Raw  json.RawMessage
}

func (*ThreadStartedEvent) threadEvent() {}
func (*TurnStartedEvent) threadEvent()   {}
func (*TurnCompletedEvent) threadEvent() {}
func (*TurnFailedEvent) threadEvent()    {}
func (*ItemStartedEvent) threadEvent()   {}
func (*ItemUpdatedEvent) threadEvent()   {}
func (*ItemCompletedEvent) threadEvent() {}
func (*ThreadErrorEvent) threadEvent()   {}
func (*UnknownEvent) threadEvent()       {}

// parseEvent decodes one JSONL line from codex stdout. Returns
// *ParseEventError on JSON failure; otherwise a typed event (possibly
// *UnknownEvent for forward-compat).
func parseEvent(line []byte) (ThreadEvent, error) {
	var head struct {
		Type string          `json:"type"`
		Item json.RawMessage `json:"item"`
	}
	if err := json.Unmarshal(line, &head); err != nil {
		return nil, &ParseEventError{Line: string(line), Err: err}
	}
	switch head.Type {
	case "thread.started":
		var e ThreadStartedEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, &ParseEventError{Line: string(line), Err: err}
		}
		return &e, nil
	case "turn.started":
		var e TurnStartedEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, &ParseEventError{Line: string(line), Err: err}
		}
		return &e, nil
	case "turn.completed":
		var e TurnCompletedEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, &ParseEventError{Line: string(line), Err: err}
		}
		return &e, nil
	case "turn.failed":
		var e TurnFailedEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, &ParseEventError{Line: string(line), Err: err}
		}
		return &e, nil
	case "item.started", "item.updated", "item.completed":
		item, err := parseItem(head.Item)
		if err != nil {
			return nil, &ParseEventError{Line: string(line), Err: err}
		}
		switch head.Type {
		case "item.started":
			return &ItemStartedEvent{Type: head.Type, Item: item, raw: head.Item}, nil
		case "item.updated":
			return &ItemUpdatedEvent{Type: head.Type, Item: item, raw: head.Item}, nil
		default:
			return &ItemCompletedEvent{Type: head.Type, Item: item, raw: head.Item}, nil
		}
	case "error":
		var e ThreadErrorEvent
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, &ParseEventError{Line: string(line), Err: err}
		}
		return &e, nil
	default:
		return &UnknownEvent{Type: head.Type, Raw: append([]byte(nil), line...)}, nil
	}
}

// parseItem stub — implemented in Task 8 (items.go). Until then, return
// UnknownItem so events.go alone compiles.
func parseItem(raw json.RawMessage) (ThreadItem, error) {
	return &UnknownItem{Raw: append([]byte(nil), raw...)}, nil
}

// Stubs replaced in Task 8.
type ThreadItem interface{ threadItem() }

type UnknownItem struct {
	Type string
	Raw  json.RawMessage
}

func (*UnknownItem) threadItem() {}
