//go:build integration

package codex

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// Run with: go test -tags=integration ./... -timeout=2m
//
// Requires: codex binary on PATH; OPENAI_API_KEY env set.
// Skipped automatically when those preconditions don't hold.

func TestIntegration_Quickstart(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	c := New(CodexOptions{APIKey: os.Getenv("OPENAI_API_KEY")})
	cwd, err := os.MkdirTemp("", "codex-itest-")
	if err != nil { t.Fatal(err) }
	defer os.RemoveAll(cwd)

	th := c.StartThread(ThreadOptions{
		SandboxMode:      SandboxReadOnly,
		WorkingDirectory: cwd,
		SkipGitRepoCheck: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	turn, err := th.Run(ctx, StringInput("Reply with the single word: pong"), TurnOptions{})
	if err != nil { t.Fatalf("Run: %v", err) }

	if !strings.Contains(strings.ToLower(turn.FinalResponse), "pong") {
		t.Errorf("FinalResponse = %q, expected to contain 'pong'", turn.FinalResponse)
	}
	if th.ID() == "" {
		t.Errorf("ID() empty after run")
	}
	if turn.Usage == nil || turn.Usage.OutputTokens == 0 {
		t.Errorf("Usage looks empty: %+v", turn.Usage)
	}
}
