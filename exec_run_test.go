package codex

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func fakeBin(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", "fake_codex", name)
}

func TestRunExec_Clean(t *testing.T) {
	stream, err := runExec(context.Background(), runExecInput{
		Binary: fakeBin(t, "clean.sh"),
		Args:   []string{"exec", "--experimental-json"},
		Env:    nil,
		Prompt: "hi",
	})
	if err != nil { t.Fatal(err) }

	var got []string
	for evt := range stream.Events() {
		switch e := evt.(type) {
		case *ThreadStartedEvent:
			got = append(got, "started:"+e.ThreadID)
		case *TurnStartedEvent:
			got = append(got, "turn.started")
		case *ItemCompletedEvent:
			if am, ok := e.Item.(*AgentMessageItem); ok {
				got = append(got, "msg:"+am.Text)
			}
		case *TurnCompletedEvent:
			got = append(got, "completed")
		}
	}
	if err := stream.Wait(); err != nil {
		t.Errorf("Wait() = %v, want nil", err)
	}

	want := []string{"started:01HMFAKE", "turn.started", "msg:hello", "completed"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Errorf("got[%d]=%q, want %q", i, got[i:], want[i:])
			break
		}
	}
}

func TestRunExec_TurnFailedYieldsButWaitNil(t *testing.T) {
	stream, err := runExec(context.Background(), runExecInput{
		Binary: fakeBin(t, "turn_failed.sh"),
		Args:   []string{"exec", "--experimental-json"},
	})
	if err != nil { t.Fatal(err) }

	sawFailed := false
	for evt := range stream.Events() {
		if _, ok := evt.(*TurnFailedEvent); ok {
			sawFailed = true
		}
	}
	if !sawFailed { t.Error("expected TurnFailedEvent on channel") }
	if err := stream.Wait(); err != nil {
		t.Errorf("Wait() = %v, want nil (turn.failed alone does not fail Wait)", err)
	}
}

func TestRunExec_NonZeroExit(t *testing.T) {
	stream, err := runExec(context.Background(), runExecInput{
		Binary: fakeBin(t, "exit_nonzero.sh"),
		Args:   []string{"exec", "--experimental-json"},
	})
	if err != nil { t.Fatal(err) }

	for range stream.Events() {} // drain
	werr := stream.Wait()
	if werr == nil {
		t.Fatal("expected error from Wait()")
	}
	var nz *NonZeroExitError
	if !errors.As(werr, &nz) {
		t.Fatalf("Wait err = %T (%v), want *NonZeroExitError", werr, werr)
	}
	if nz.Code != 7 {
		t.Errorf("Code = %d", nz.Code)
	}
	if nz.Stderr == "" {
		t.Errorf("Stderr should be populated, got empty")
	}
}

func TestRunExec_MalformedLineYieldsErrorEventNotWaitErr(t *testing.T) {
	stream, err := runExec(context.Background(), runExecInput{
		Binary: fakeBin(t, "malformed.sh"),
		Args:   []string{"exec", "--experimental-json"},
	})
	if err != nil { t.Fatal(err) }

	sawErr := false
	sawCompleted := false
	for evt := range stream.Events() {
		if _, ok := evt.(*ThreadErrorEvent); ok {
			sawErr = true
		}
		if _, ok := evt.(*TurnCompletedEvent); ok {
			sawCompleted = true
		}
	}
	if !sawErr      { t.Error("expected synthetic ThreadErrorEvent for malformed line") }
	if !sawCompleted { t.Error("expected scanner to continue past bad line") }
	if err := stream.Wait(); err != nil {
		t.Errorf("Wait() = %v, want nil", err)
	}
}

func TestRunExec_CtxCancelEscalatesToKill(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := runExec(ctx, runExecInput{
		Binary: fakeBin(t, "hang.sh"),
		Args:   []string{"exec", "--experimental-json"},
	})
	if err != nil { t.Fatal(err) }

	saw := 0
	go func() {
		for range stream.Events() {
			saw++
			if saw == 1 {
				cancel()
			}
		}
	}()

	start := time.Now()
	werr := stream.Wait()
	elapsed := time.Since(start)
	if !errors.Is(werr, context.Canceled) {
		var nz *NonZeroExitError
		if werr == nil || (!errors.As(werr, &nz) && !errors.Is(werr, context.Canceled)) {
			t.Errorf("Wait() = %v, want ctx.Canceled or NonZeroExitError", werr)
		}
	}
	if elapsed > 5*time.Second {
		t.Errorf("Wait took %v, expected <5s with SIGKILL escalation", elapsed)
	}
}
