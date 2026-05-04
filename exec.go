package codex

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// buildArgsInput collects all inputs to buildArgs for clean test wiring.
type buildArgsInput struct {
	CodexOpts        CodexOptions
	ThreadOpts       ThreadOptions
	ThreadID         string   // empty = no resume
	Images           []string // from joinTextParts
	OutputSchemaPath string   // empty = no --output-schema
}

// buildArgs constructs the argv for `codex exec`. The order mirrors TS
// exec.ts:73-148 line-for-line so behavior is bit-identical (modulo the
// listed divergences).
func buildArgs(in buildArgsInput) ([]string, error) {
	args := []string{"exec", "--experimental-json"}

	// 1. CodexOptions.Config — flatten and apply BEFORE per-thread flags
	overrides, err := serializeConfigOverrides(in.CodexOpts.Config)
	if err != nil {
		return nil, err
	}
	for _, o := range overrides {
		args = append(args, "--config", o)
	}

	// 2. baseUrl
	if in.CodexOpts.BaseURL != "" {
		quoted, _ := tomlValue(in.CodexOpts.BaseURL, "openai_base_url")
		args = append(args, "--config", "openai_base_url="+quoted)
	}

	// 3. model / sandbox / cwd / additional dirs / skip-git
	if in.ThreadOpts.Model != "" {
		args = append(args, "--model", in.ThreadOpts.Model)
	}
	if in.ThreadOpts.SandboxMode != "" {
		args = append(args, "--sandbox", string(in.ThreadOpts.SandboxMode))
	}
	if in.ThreadOpts.WorkingDirectory != "" {
		args = append(args, "--cd", in.ThreadOpts.WorkingDirectory)
	}
	for _, d := range in.ThreadOpts.AdditionalDirectories {
		args = append(args, "--add-dir", d)
	}
	if in.ThreadOpts.SkipGitRepoCheck {
		args = append(args, "--skip-git-repo-check")
	}

	// 4. output-schema (after fs paths, before reasoning/web/approval)
	if in.OutputSchemaPath != "" {
		args = append(args, "--output-schema", in.OutputSchemaPath)
	}

	// 5. reasoning
	if in.ThreadOpts.ModelReasoningEffort != "" {
		args = append(args, "--config", `model_reasoning_effort="`+string(in.ThreadOpts.ModelReasoningEffort)+`"`)
	}

	// 6. network access
	if in.ThreadOpts.NetworkAccessEnabled != nil {
		v := "false"
		if *in.ThreadOpts.NetworkAccessEnabled {
			v = "true"
		}
		args = append(args, "--config", "sandbox_workspace_write.network_access="+v)
	}

	// 7. web search (mode wins over legacy enabled)
	switch {
	case in.ThreadOpts.WebSearchMode != "":
		args = append(args, "--config", `web_search="`+string(in.ThreadOpts.WebSearchMode)+`"`)
	case in.ThreadOpts.WebSearchEnabled != nil && *in.ThreadOpts.WebSearchEnabled:
		args = append(args, "--config", `web_search="live"`)
	case in.ThreadOpts.WebSearchEnabled != nil && !*in.ThreadOpts.WebSearchEnabled:
		args = append(args, "--config", `web_search="disabled"`)
	}

	// 8. approval policy
	if in.ThreadOpts.ApprovalPolicy != "" {
		args = append(args, "--config", `approval_policy="`+string(in.ThreadOpts.ApprovalPolicy)+`"`)
	}

	// 9. resume subcommand (must come AFTER exec flags, BEFORE images)
	if in.ThreadID != "" {
		args = append(args, "resume", in.ThreadID)
	}

	// 10. images (parsed by `resume` subcommand, OR by `exec` if no resume)
	for _, img := range in.Images {
		args = append(args, "--image", img)
	}

	return args, nil
}

// composeEnv mirrors TS exec.ts:148-167. Returned slice is in
// "KEY=VALUE" form ready for cmd.Env.
//
// procEnv is normally os.Environ(); accepted as a parameter for testability.
func composeEnv(opts CodexOptions, procEnv []string) []string {
	env := map[string]string{}
	if opts.Env != nil {
		for k, v := range opts.Env {
			env[k] = v
		}
	} else {
		for _, kv := range procEnv {
			eq := strings.IndexByte(kv, '=')
			if eq < 0 {
				continue
			}
			env[kv[:eq]] = kv[eq+1:]
		}
	}
	if env["CODEX_INTERNAL_ORIGINATOR_OVERRIDE"] == "" {
		env["CODEX_INTERNAL_ORIGINATOR_OVERRIDE"] = "codex_sdk_go"
	}
	if opts.APIKey != "" {
		env["CODEX_API_KEY"] = opts.APIKey
	}
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

type runExecInput struct {
	Binary string
	Args   []string
	Env    []string // KEY=VALUE form; nil = inherit from os/exec default (current process)
	Prompt string   // written to stdin then closed
}

// stream is the internal *StreamedTurn that runExec returns.
type stream struct {
	events      chan ThreadEvent
	waitDone    chan struct{}
	terminalErr error
}

func (s *stream) Events() <-chan ThreadEvent { return s.events }
func (s *stream) Wait() error {
	<-s.waitDone
	return s.terminalErr
}

const (
	stderrCap     = 64 * 1024
	scannerBufMax = 4 * 1024 * 1024
)

// runExec spawns the codex process, pipes prompt to stdin, parses JSONL on
// stdout into the events channel, captures stderr for diagnostics, and
// returns a *stream whose Wait() reports terminal status.
//
// Cancellation: SIGTERM on ctx.Done, escalating to SIGKILL after 2s via
// cmd.WaitDelay (Go 1.20+).
func runExec(ctx context.Context, in runExecInput) (*stream, error) {
	cmd := exec.CommandContext(ctx, in.Binary, in.Args...)
	cmd.Env = in.Env
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 2 * time.Second

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, &SpawnError{Err: fmt.Errorf("stdin pipe: %w", err)}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, &SpawnError{Err: fmt.Errorf("stdout pipe: %w", err)}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, &SpawnError{Err: fmt.Errorf("stderr pipe: %w", err)}
	}

	if err := cmd.Start(); err != nil {
		return nil, &SpawnError{Err: err}
	}

	// Write prompt then close stdin.
	go func() {
		_, _ = io.WriteString(stdin, in.Prompt)
		_ = stdin.Close()
	}()

	// Drain stderr into a bounded buffer.
	var stderrBuf bytes.Buffer
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		buf := make([]byte, 4096)
		for {
			n, rerr := stderr.Read(buf)
			if n > 0 {
				if stderrBuf.Len() < stderrCap {
					space := stderrCap - stderrBuf.Len()
					if n > space {
						n = space
					}
					stderrBuf.Write(buf[:n])
					if stderrBuf.Len() == stderrCap {
						stderrBuf.WriteString("...[truncated]\n")
					}
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	// waitResult carries the return value of cmd.Wait(). cmd.Wait() must be
	// called promptly after Start() so that the WaitDelay timer (which
	// triggers SIGKILL escalation) is armed by the Go runtime. If we delay
	// calling cmd.Wait() until after the stdout scanner finishes we create a
	// deadlock: the scanner blocks waiting for the pipe to close, the pipe
	// only closes when the process exits or when WaitDelay fires, and
	// WaitDelay never fires because cmd.Wait() hasn't been called yet.
	waitResult := make(chan error, 1)
	go func() { waitResult <- cmd.Wait() }()

	s := &stream{
		events:   make(chan ThreadEvent, 16),
		waitDone: make(chan struct{}),
	}

	go func() {
		// Defers run LIFO. We MUST close events BEFORE waitDone so that
		// Wait() unblocking implies the events channel is fully closed
		// (per spec contract). Register waitDone first so it runs last.
		defer close(s.waitDone)
		defer close(s.events)

		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 0, 64*1024), scannerBufMax)
		var parseFatal error
		for sc.Scan() {
			line := sc.Bytes()
			if len(strings.TrimSpace(string(line))) == 0 {
				continue
			}
			evt, perr := parseEvent(line)
			if perr != nil {
				// TS thread.ts:99-103 throws on JSON.parse failure, which
				// terminates the generator with that error. We mirror by
				// killing the subprocess, setting terminalErr, and exiting
				// the scanner loop (no synthetic ThreadErrorEvent).
				parseFatal = perr
				_ = cmd.Process.Kill()
				break
			}
			s.events <- evt
		}
		// Scanner exited (EOF, error, or parse-fatal). Drain stderr fully.
		<-stderrDone

		if parseFatal != nil {
			s.terminalErr = parseFatal
			// Reap the subprocess (already killed) so the wait goroutine exits.
			<-waitResult
			return
		}

		// Collect the result from cmd.Wait() (already running in parallel).
		werr := <-waitResult
		if werr == nil {
			return // clean exit, terminalErr stays nil
		}

		if ctx.Err() != nil {
			s.terminalErr = ctx.Err()
			s.events <- &ThreadErrorEvent{Type: "error", Message: "codex cancelled: " + ctx.Err().Error()}
			return
		}

		var ee *exec.ExitError
		if errors.As(werr, &ee) {
			ws, _ := ee.Sys().(syscall.WaitStatus)
			code := ee.ExitCode()
			signal := ""
			if ws.Signaled() {
				signal = ws.Signal().String()
			}
			nz := &NonZeroExitError{Code: code, Signal: signal, Stderr: stderrBuf.String()}
			s.terminalErr = nz
			s.events <- &ThreadErrorEvent{Type: "error", Message: nz.Error()}
			return
		}

		// Other Wait error.
		s.terminalErr = werr
		s.events <- &ThreadErrorEvent{Type: "error", Message: werr.Error()}
	}()

	return s, nil
}
