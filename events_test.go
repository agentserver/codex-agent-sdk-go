package codex

import (
	"testing"
)

func TestParseEvent_ThreadStarted(t *testing.T) {
	line := `{"type":"thread.started","thread_id":"01HM..."}`
	evt, err := parseEvent([]byte(line))
	if err != nil { t.Fatal(err) }
	ts, ok := evt.(*ThreadStartedEvent)
	if !ok { t.Fatalf("got %T", evt) }
	if ts.ThreadID != "01HM..." { t.Errorf("ThreadID = %q", ts.ThreadID) }
}

func TestParseEvent_TurnStarted(t *testing.T) {
	evt, err := parseEvent([]byte(`{"type":"turn.started"}`))
	if err != nil { t.Fatal(err) }
	if _, ok := evt.(*TurnStartedEvent); !ok {
		t.Fatalf("got %T", evt)
	}
}

func TestParseEvent_TurnCompleted(t *testing.T) {
	line := `{"type":"turn.completed","usage":{"input_tokens":10,"cached_input_tokens":5,"output_tokens":20,"reasoning_output_tokens":3}}`
	evt, err := parseEvent([]byte(line))
	if err != nil { t.Fatal(err) }
	tc, ok := evt.(*TurnCompletedEvent)
	if !ok { t.Fatalf("got %T", evt) }
	want := Usage{InputTokens: 10, CachedInputTokens: 5, OutputTokens: 20, ReasoningOutputTokens: 3}
	if tc.Usage != want { t.Errorf("Usage = %+v want %+v", tc.Usage, want) }
}

func TestParseEvent_TurnFailed(t *testing.T) {
	line := `{"type":"turn.failed","error":{"message":"model rejected"}}`
	evt, err := parseEvent([]byte(line))
	if err != nil { t.Fatal(err) }
	tf, ok := evt.(*TurnFailedEvent)
	if !ok { t.Fatalf("got %T", evt) }
	if tf.Error.Message != "model rejected" { t.Errorf("msg = %q", tf.Error.Message) }
}

func TestParseEvent_ThreadError(t *testing.T) {
	evt, err := parseEvent([]byte(`{"type":"error","message":"explode"}`))
	if err != nil { t.Fatal(err) }
	te, ok := evt.(*ThreadErrorEvent)
	if !ok { t.Fatalf("got %T", evt) }
	if te.Message != "explode" { t.Errorf("msg = %q", te.Message) }
}

func TestParseEvent_UnknownType(t *testing.T) {
	evt, err := parseEvent([]byte(`{"type":"future.event","x":1}`))
	if err != nil { t.Fatal(err) }
	u, ok := evt.(*UnknownEvent)
	if !ok { t.Fatalf("got %T", evt) }
	if u.Type != "future.event" { t.Errorf("Type = %q", u.Type) }
	if string(u.Raw) == "" { t.Error("Raw should be populated") }
}

func TestParseEvent_Malformed(t *testing.T) {
	_, err := parseEvent([]byte(`{not json`))
	if err == nil { t.Error("expected error for malformed JSON") }
	pe, ok := err.(*ParseEventError)
	if !ok { t.Fatalf("got %T", err) }
	if pe.Line == "" { t.Error("Line should be populated") }
}

func TestParseEvent_ItemDelegated(t *testing.T) {
	line := `{"type":"item.started","item":{"id":"i1","type":"reasoning","text":"think"}}`
	evt, err := parseEvent([]byte(line))
	if err != nil { t.Fatal(err) }
	is, ok := evt.(*ItemStartedEvent)
	if !ok { t.Fatalf("got %T", evt) }
	if is.Item == nil { t.Error("Item should be non-nil") }
}
