package codex

// Codex is the top-level handle. Cheap to construct; safe to share across
// goroutines (state lives on Thread, not Codex).
type Codex struct {
	opts CodexOptions
}

// New creates a Codex. CodexPathOverride defaults to "codex" (PATH lookup).
func New(opts CodexOptions) *Codex {
	if opts.CodexPathOverride == "" {
		opts.CodexPathOverride = "codex"
	}
	return &Codex{opts: opts}
}

// StartThread builds a Thread with no thread_id. The first RunStreamed
// captures the codex-generated id from `thread.started`.
func (c *Codex) StartThread(topts ThreadOptions) *Thread {
	return &Thread{codex: c, topts: topts}
}

// ResumeThread builds a Thread already bound to threadID. Every
// RunStreamed appends `resume <threadID>` to the codex invocation.
func (c *Codex) ResumeThread(threadID string, topts ThreadOptions) *Thread {
	return &Thread{codex: c, topts: topts, id: threadID}
}
