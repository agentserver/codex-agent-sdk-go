package codex

import "fmt"

// SpawnError wraps a failure from os/exec before the codex process starts
// (e.g., binary not on PATH, permission denied).
type SpawnError struct{ Err error }

func (e *SpawnError) Error() string { return "codex spawn: " + e.Err.Error() }
func (e *SpawnError) Unwrap() error { return e.Err }

// NonZeroExitError is returned when codex exits with a non-zero code or is
// killed by a signal. Stderr is the bounded tail of the subprocess's
// stderr stream (capped at 64KB). Underlying preserves the original
// *exec.ExitError for callers who want errors.As(err, new(*exec.ExitError)).
type NonZeroExitError struct {
	Code       int
	Signal     string
	Stderr     string
	Underlying error // typically *exec.ExitError; may be nil if not available
}

func (e *NonZeroExitError) Error() string {
	if e.Signal != "" {
		return fmt.Sprintf("codex exited with signal %s: %s", e.Signal, e.Stderr)
	}
	return fmt.Sprintf("codex exited with code %d: %s", e.Code, e.Stderr)
}

// Unwrap exposes the underlying *exec.ExitError (if present) so callers can
// use errors.Is / errors.As to inspect OS-level exit details.
func (e *NonZeroExitError) Unwrap() error { return e.Underlying }

// ParseEventError indicates a JSONL line from codex stdout could not be
// parsed into a known event type. ParseEventError does NOT terminate the
// stream; it is wrapped into a synthetic ThreadErrorEvent.
type ParseEventError struct {
	Line string
	Err  error
}

func (e *ParseEventError) Error() string {
	return fmt.Sprintf("parse event: %v (line: %q)", e.Err, truncate(e.Line, 120))
}
func (e *ParseEventError) Unwrap() error { return e.Err }

// TurnFailedError is returned by Thread.Run when codex emits a turn.failed
// event. Thread.RunStreamed does not return this — it yields the
// TurnFailedEvent on the channel and lets the caller decide.
type TurnFailedError struct{ Message string }

func (e *TurnFailedError) Error() string { return "codex turn failed: " + e.Message }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
