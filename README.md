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
