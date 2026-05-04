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
