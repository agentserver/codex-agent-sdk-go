package main

import (
	"context"
	"fmt"
	"os"

	codex "github.com/agentserver/codex-agent-sdk-go"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY not set")
		os.Exit(1)
	}

	c := codex.New(codex.CodexOptions{APIKey: apiKey})

	cwd, err := os.MkdirTemp("", "codex-quickstart-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(cwd)

	th := c.StartThread(codex.ThreadOptions{
		SandboxMode:      codex.SandboxReadOnly,
		WorkingDirectory: cwd,
		SkipGitRepoCheck: true,
	})

	stream, err := th.RunStreamed(context.Background(),
		codex.StringInput("Tell me a one-line haiku about Go."),
		codex.TurnOptions{})
	if err != nil {
		panic(err)
	}

	for evt := range stream.Events() {
		switch e := evt.(type) {
		case *codex.ItemCompletedEvent:
			if am, ok := e.Item.(*codex.AgentMessageItem); ok {
				fmt.Println("AGENT:", am.Text)
			}
		case *codex.TurnCompletedEvent:
			fmt.Printf("USAGE: in=%d out=%d\n", e.Usage.InputTokens, e.Usage.OutputTokens)
		}
	}
	if err := stream.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "stream error:", err)
		os.Exit(1)
	}
	fmt.Println("THREAD ID:", th.ID())
}
