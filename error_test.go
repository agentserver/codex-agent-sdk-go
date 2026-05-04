package codex

import (
	"errors"
	"strings"
	"testing"
)

func TestSpawnError(t *testing.T) {
	inner := errors.New("no such file")
	e := &SpawnError{Err: inner}
	if !strings.Contains(e.Error(), "no such file") {
		t.Errorf("Error() = %q, want it to contain %q", e.Error(), "no such file")
	}
	if !errors.Is(e, inner) {
		t.Errorf("errors.Is should unwrap to inner")
	}
}

func TestNonZeroExitError(t *testing.T) {
	e := &NonZeroExitError{Code: 2, Signal: "", Stderr: "boom"}
	got := e.Error()
	for _, want := range []string{"code 2", "boom"} {
		if !strings.Contains(got, want) {
			t.Errorf("Error() = %q, missing %q", got, want)
		}
	}

	e2 := &NonZeroExitError{Code: -1, Signal: "killed", Stderr: ""}
	if !strings.Contains(e2.Error(), "signal killed") {
		t.Errorf("Error() with signal = %q", e2.Error())
	}
}

func TestParseEventError(t *testing.T) {
	inner := errors.New("unexpected token")
	e := &ParseEventError{Line: "{garbage", Err: inner}
	if !strings.Contains(e.Error(), "unexpected token") {
		t.Errorf("Error() = %q", e.Error())
	}
	if !errors.Is(e, inner) {
		t.Errorf("errors.Is should unwrap")
	}
}

func TestTurnFailedError(t *testing.T) {
	e := &TurnFailedError{Message: "model rejected"}
	if e.Error() != "codex turn failed: model rejected" {
		t.Errorf("Error() = %q", e.Error())
	}
}
