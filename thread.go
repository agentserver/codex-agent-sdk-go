package codex

import (
	"context"
	"os"
	"sync"
)

// Thread mirrors TS `Thread`. NOT safe for concurrent Run / RunStreamed
// — codex CLI is one-prompt-per-process; serialize at the caller level.
type Thread struct {
	codex *Codex
	topts ThreadOptions

	idMu sync.RWMutex
	id   string
}

// ID returns the thread_id, captured from the first thread.started event
// or set explicitly by ResumeThread. Returns "" before either has happened.
func (t *Thread) ID() string {
	t.idMu.RLock()
	defer t.idMu.RUnlock()
	return t.id
}

func (t *Thread) setID(id string) {
	t.idMu.Lock()
	t.id = id
	t.idMu.Unlock()
}

// StreamedTurn mirrors TS `RunStreamedResult`. Events() yields typed
// events in order; Wait() returns once the events channel is fully
// drained AND the codex subprocess has exited.
type StreamedTurn = stream

// RunStreamedResult is a TS-parity alias for StreamedTurn.
type RunStreamedResult = StreamedTurn

// RunStreamed sends the prompt and returns a stream of typed events.
func (t *Thread) RunStreamed(ctx context.Context, input Input, topts TurnOptions) (*StreamedTurn, error) {
	prompt, images := joinTextParts(input)

	schemaPath, cleanupSchema, err := prepareOutputSchema(topts.OutputSchema)
	if err != nil {
		return nil, err
	}

	args, err := buildArgs(buildArgsInput{
		CodexOpts:        t.codex.opts,
		ThreadOpts:       t.topts,
		ThreadID:         t.ID(),
		Images:           images,
		OutputSchemaPath: schemaPath,
	})
	if err != nil {
		cleanupSchema()
		return nil, err
	}

	env := composeEnv(t.codex.opts, os.Environ())

	s, err := runExec(ctx, runExecInput{
		Binary: t.codex.opts.BinaryPath,
		Args:   args,
		Env:    env,
		Prompt: prompt,
	})
	if err != nil {
		cleanupSchema()
		return nil, err
	}

	wrapped := wrapStream(s, t, cleanupSchema)
	return wrapped, nil
}

// wrapStream tees the inner stream's events through a goroutine that
// intercepts ThreadStartedEvent (to set Thread.id atomically before
// forwarding) and ensures cleanupSchema runs after Wait completes.
func wrapStream(inner *stream, t *Thread, cleanupSchema func()) *stream {
	out := &stream{
		events:   make(chan ThreadEvent, 16),
		waitDone: make(chan struct{}),
	}
	go func() {
		// LIFO: events must close before waitDone (see runExec for rationale).
		defer close(out.waitDone)
		defer close(out.events)
		for evt := range inner.Events() {
			if started, ok := evt.(*ThreadStartedEvent); ok {
				t.setID(started.ThreadID)
			}
			out.events <- evt
		}
		out.terminalErr = inner.Wait()
		cleanupSchema()
	}()
	return out
}

// Turn mirrors TS `RunResult`. Aggregated state from a buffered Run().
type Turn struct {
	Items         []ThreadItem
	FinalResponse string
	Usage         *Usage
}

// RunResult is a TS-parity alias for Turn.
type RunResult = Turn

// Run is the buffered convenience wrapper around RunStreamed. Mirrors TS
// `Thread.run` (thread.ts:115-138):
//
//   - item.completed → appended to Items; AgentMessageItem.Text overwrites FinalResponse
//   - turn.completed → Usage set
//   - turn.failed   → returns (zero Turn, *TurnFailedError); channel still drained
//   - subprocess errors (Spawn/NonZeroExit/ctx) → returned as-is from Wait()
func (t *Thread) Run(ctx context.Context, input Input, topts TurnOptions) (Turn, error) {
	stream, err := t.RunStreamed(ctx, input, topts)
	if err != nil { return Turn{}, err }

	var turn Turn
	var failed *TurnFailedError
	for evt := range stream.Events() {
		switch e := evt.(type) {
		case *ItemCompletedEvent:
			turn.Items = append(turn.Items, e.Item)
			if am, ok := e.Item.(*AgentMessageItem); ok {
				turn.FinalResponse = am.Text
			}
		case *TurnCompletedEvent:
			u := e.Usage
			turn.Usage = &u
		case *TurnFailedEvent:
			msg := e.Error.Message
			failed = &TurnFailedError{Message: msg}
		}
	}

	if werr := stream.Wait(); werr != nil {
		return Turn{}, werr
	}
	if failed != nil {
		return Turn{}, failed
	}
	return turn, nil
}
