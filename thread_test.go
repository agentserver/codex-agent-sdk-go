package codex

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fakeCodex(t *testing.T, script string) *Codex {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	bin := filepath.Join(filepath.Dir(thisFile), "testdata", "fake_codex", script)
	return New(CodexOptions{CodexPathOverride: bin})
}

func TestStartThread_CapturesIDOnFirstStartedEvent(t *testing.T) {
	c := fakeCodex(t, "clean.sh")
	th := c.StartThread(ThreadOptions{})
	if th.ID() != "" {
		t.Errorf("ID before run = %q, want empty", th.ID())
	}

	stream, err := th.RunStreamed(context.Background(), StringInput("hi"), TurnOptions{})
	if err != nil { t.Fatal(err) }
	for range stream.Events() {} // drain
	if err := stream.Wait(); err != nil { t.Fatal(err) }

	if th.ID() != "01HMFAKE" {
		t.Errorf("ID after run = %q, want 01HMFAKE", th.ID())
	}
}

func TestResumeThread_PassesResumeArg(t *testing.T) {
	bin := writeArgEchoScript(t)
	c := New(CodexOptions{CodexPathOverride: bin})
	th := c.ResumeThread("preset-id", ThreadOptions{})
	if th.ID() != "preset-id" {
		t.Errorf("ID before run = %q", th.ID())
	}

	stream, err := th.RunStreamed(context.Background(), StringInput("hi"), TurnOptions{})
	if err != nil { t.Fatal(err) }

	var msg string
	for evt := range stream.Events() {
		if ic, ok := evt.(*ItemCompletedEvent); ok {
			if am, ok := ic.Item.(*AgentMessageItem); ok {
				msg = am.Text
			}
		}
	}
	if err := stream.Wait(); err != nil { t.Fatal(err) }

	if !contains(msg, "resume preset-id") {
		t.Errorf("args echo = %q, missing 'resume preset-id'", msg)
	}
}

func TestStartThread_SecondRunSwitchesToResume(t *testing.T) {
	bin := writeArgEchoScript(t)
	c := New(CodexOptions{CodexPathOverride: bin})
	th := c.StartThread(ThreadOptions{})

	s1, err := th.RunStreamed(context.Background(), StringInput("first"), TurnOptions{})
	if err != nil { t.Fatal(err) }
	for range s1.Events() {}
	if err := s1.Wait(); err != nil { t.Fatal(err) }
	if th.ID() != "echoed-id" {
		t.Fatalf("first-run ID = %q, want echoed-id", th.ID())
	}

	s2, err := th.RunStreamed(context.Background(), StringInput("second"), TurnOptions{})
	if err != nil { t.Fatal(err) }
	var msg2 string
	for evt := range s2.Events() {
		if ic, ok := evt.(*ItemCompletedEvent); ok {
			if am, ok := ic.Item.(*AgentMessageItem); ok {
				msg2 = am.Text
			}
		}
	}
	if err := s2.Wait(); err != nil { t.Fatal(err) }

	if !contains(msg2, "resume echoed-id") {
		t.Errorf("2nd-run args = %q, missing resume", msg2)
	}
}

// writeArgEchoScript creates a one-off bash fake that emits its argv as an
// agent_message item, plus a thread.started with id "echoed-id".
func writeArgEchoScript(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "echo.sh")
	body := `#!/bin/bash
cat > /dev/null
echo '{"type":"thread.started","thread_id":"echoed-id"}'
ARGS_JSON=$(printf '%s ' "$@" | sed 's/"/\\"/g')
echo "{\"type\":\"item.completed\",\"item\":{\"id\":\"i1\",\"type\":\"agent_message\",\"text\":\"args: ${ARGS_JSON}\"}}"
echo '{"type":"turn.completed","usage":{"input_tokens":0,"cached_input_tokens":0,"output_tokens":0,"reasoning_output_tokens":0}}'
exit 0
`
	if err := os.WriteFile(bin, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
