// Package codex is a Go SDK for the codex CLI binary, a Go port of
// @openai/codex-sdk (TypeScript). It spawns `codex exec --experimental-json`
// per turn, writes the prompt to stdin, and yields typed events parsed
// from stdout JSONL.
//
// See README.md for a quickstart and the design spec at
// docs/superpowers/specs/2026-05-04-codex-agent-sdk-go-design.md (in the
// agentserver repo) for full alignment notes vs the TS SDK.
package codex
