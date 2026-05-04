package codex

import (
	"context"
	"errors"
	"testing"
)

func TestRun_BuffersUntilCompletion(t *testing.T) {
	c := fakeCodex(t, "clean.sh")
	th := c.StartThread(ThreadOptions{})
	turn, err := th.Run(context.Background(), StringInput("hi"), TurnOptions{})
	if err != nil { t.Fatal(err) }
	if turn.FinalResponse != "hello" {
		t.Errorf("FinalResponse = %q", turn.FinalResponse)
	}
	if len(turn.Items) != 1 {
		t.Errorf("Items = %d, want 1", len(turn.Items))
	}
	if turn.Usage == nil || turn.Usage.OutputTokens != 2 {
		t.Errorf("Usage = %+v", turn.Usage)
	}
	if th.ID() != "01HMFAKE" {
		t.Errorf("ID = %q", th.ID())
	}
}

func TestRun_TurnFailedReturnsTurnFailedError(t *testing.T) {
	c := fakeCodex(t, "turn_failed.sh")
	th := c.StartThread(ThreadOptions{})
	_, err := th.Run(context.Background(), StringInput("hi"), TurnOptions{})
	if err == nil { t.Fatal("expected error") }
	var tf *TurnFailedError
	if !errors.As(err, &tf) {
		t.Fatalf("err = %T (%v), want *TurnFailedError", err, err)
	}
	if tf.Message != "model rejected" {
		t.Errorf("Message = %q", tf.Message)
	}
}

func TestRun_NonZeroExitReturnsExitError(t *testing.T) {
	c := fakeCodex(t, "exit_nonzero.sh")
	th := c.StartThread(ThreadOptions{})
	_, err := th.Run(context.Background(), StringInput("hi"), TurnOptions{})
	if err == nil { t.Fatal("expected error") }
	var nz *NonZeroExitError
	if !errors.As(err, &nz) {
		t.Fatalf("err = %T (%v), want *NonZeroExitError", err, err)
	}
}
