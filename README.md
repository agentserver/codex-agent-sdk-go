# codex-agent-sdk-go

Go SDK for the [codex](https://github.com/openai/codex) CLI, ported from
`@openai/codex-sdk` (TypeScript).

Wraps `codex exec --experimental-json`: spawns the CLI per turn, writes
the prompt to stdin, parses JSONL events from stdout into typed Go events.

## Status

Pre-1.0. API tracks the TS SDK exactly; see the design spec for the
complete alignment list.

## Requirements

- Go 1.22+
- `codex` CLI binary on PATH (or pass `CodexOptions.BinaryPath`)
- Tested against codex-cli >= 0.125.0

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/agentserver/codex-agent-sdk-go"
)

func main() {
    c := codex.New(codex.CodexOptions{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    t := c.StartThread(codex.ThreadOptions{
        SandboxMode:      codex.SandboxWorkspaceWrite,
        WorkingDirectory: "/tmp/work",
        SkipGitRepoCheck: true,
    })
    turn, err := t.Run(context.Background(),
        codex.StringInput("List the files in this directory"),
        codex.TurnOptions{})
    if err != nil { panic(err) }
    fmt.Println(turn.FinalResponse)
    fmt.Println("thread id:", t.ID())
}
```

## Streaming events

```go
stream, _ := th.RunStreamed(ctx, codex.StringInput("..."), codex.TurnOptions{})
for evt := range stream.Events() {
    switch e := evt.(type) {
    case *codex.ItemCompletedEvent:
        if am, ok := e.Item.(*codex.AgentMessageItem); ok {
            fmt.Println(am.Text)
        }
    case *codex.TurnFailedEvent:
        fmt.Println("turn failed:", e.Error.Message)
    }
}
if err := stream.Wait(); err != nil {
    log.Fatal(err)
}
```

## Resume

```go
// Pre-existing thread (from a prior session):
th := c.ResumeThread("01HMTHREAD...", codex.ThreadOptions{})

// Or capture id from a fresh thread:
th := c.StartThread(codex.ThreadOptions{})
turn, _ := th.Run(ctx, codex.StringInput("..."), codex.TurnOptions{})
fmt.Println(th.ID()) // populated after first turn
turn2, _ := th.Run(ctx, codex.StringInput("continue"), codex.TurnOptions{})
// turn2 implicitly resumes — the SDK appends `resume <id>` automatically.
```

## Errors

- `*codex.SpawnError` — codex binary couldn't start (PATH, perms)
- `*codex.NonZeroExitError` — codex exited non-zero or by signal
- `*codex.ParseEventError` — JSONL line failed to parse (also surfaces as
  a synthetic `ThreadErrorEvent` on the channel)
- `*codex.TurnFailedError` — `Run()` only; `RunStreamed` yields the
  `TurnFailedEvent` instead

## Alignment with `@openai/codex-sdk` (TypeScript)

This SDK is a port of the official TypeScript SDK. Every option, default,
env var, and CLI argument is reproduced. Intentional divergences:

1. **Binary discovery** — PATH lookup only (no npm platform-package
   fallback). Override via `CodexOptions.BinaryPath`.
2. **Cancellation** — `ctx.Done` triggers SIGTERM with 2s grace, then
   SIGKILL. TS uses single SIGTERM.
3. **Originator** — `CODEX_INTERNAL_ORIGINATOR_OVERRIDE` defaults to
   `"codex_sdk_go"`.
4. **Concurrency** — `Thread.Run` / `RunStreamed` are documented as
   not safe for concurrent calls (TS doesn't formalize this).

See full design at
`docs/superpowers/specs/2026-05-04-codex-agent-sdk-go-design.md` in the
agentserver repo.
