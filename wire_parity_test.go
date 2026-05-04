package codex

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func runSpy(t *testing.T, c *Codex, th *Thread, input Input, opts TurnOptions) (argv []string, env map[string]string, stdin string) {
	t.Helper()
	out := t.TempDir()
	t.Setenv("SPY_OUT", out)

	stream, err := th.RunStreamed(context.Background(), input, opts)
	if err != nil {
		t.Fatal(err)
	}
	for range stream.Events() {
	}
	if err := stream.Wait(); err != nil {
		t.Fatal(err)
	}

	argvBytes, err := os.ReadFile(filepath.Join(out, "argv"))
	if err != nil {
		t.Fatal(err)
	}
	envBytes, err := os.ReadFile(filepath.Join(out, "env"))
	if err != nil {
		t.Fatal(err)
	}
	stdinBytes, err := os.ReadFile(filepath.Join(out, "stdin"))
	if err != nil {
		t.Fatal(err)
	}

	argv = strings.Split(strings.TrimRight(string(argvBytes), "\n"), "\n")
	env = map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(string(envBytes), "\n"), "\n") {
		if i := strings.IndexByte(line, '='); i >= 0 {
			env[line[:i]] = line[i+1:]
		}
	}
	stdin = string(stdinBytes)
	return
}

func spyCodex(t *testing.T, opts CodexOptions) *Codex {
	_, thisFile, _, _ := runtime.Caller(0)
	opts.BinaryPath = filepath.Join(filepath.Dir(thisFile), "testdata", "fake_codex", "spy.sh")
	return New(opts)
}

func TestWireParity_StartThread_MinimalArgs(t *testing.T) {
	c := spyCodex(t, CodexOptions{APIKey: "sk-test"})
	th := c.StartThread(ThreadOptions{})
	argv, env, stdin := runSpy(t, c, th, StringInput("hello"), TurnOptions{})

	wantArgv := []string{"exec", "--experimental-json"}
	if !reflect.DeepEqual(argv, wantArgv) {
		t.Errorf("argv = %v, want %v", argv, wantArgv)
	}
	if env["CODEX_API_KEY"] != "sk-test" {
		t.Errorf("CODEX_API_KEY = %q", env["CODEX_API_KEY"])
	}
	if env["CODEX_INTERNAL_ORIGINATOR_OVERRIDE"] != "codex_sdk_go" {
		t.Errorf("originator = %q", env["CODEX_INTERNAL_ORIGINATOR_OVERRIDE"])
	}
	if stdin != "hello" {
		t.Errorf("stdin = %q, want %q", stdin, "hello")
	}
}

func TestWireParity_FullArgs_MatchesSpecOrdering(t *testing.T) {
	tt := true
	c := spyCodex(t, CodexOptions{
		BaseURL: "https://api.x.example/v1",
		APIKey:  "sk-x",
	})
	th := c.StartThread(ThreadOptions{
		Model:                "o3",
		SandboxMode:          SandboxWorkspaceWrite,
		WorkingDirectory:     "/tmp/w",
		AdditionalDirs:       []string{"/d1"},
		SkipGitRepoCheck:     true,
		ModelReasoningEffort: ReasoningHigh,
		NetworkAccessEnabled: &tt,
		WebSearchMode:        WebSearchLive,
		ApprovalPolicy:       ApprovalOnRequest,
	})
	argv, _, _ := runSpy(t, c, th, PartsInput{
		{Type: InputText, Text: "hello"},
		{Type: InputLocalImage, Path: "/x.png"},
	}, TurnOptions{})

	want := []string{
		"exec", "--experimental-json",
		"--config", `openai_base_url="https://api.x.example/v1"`,
		"--model", "o3",
		"--sandbox", "workspace-write",
		"--cd", "/tmp/w",
		"--add-dir", "/d1",
		"--skip-git-repo-check",
		"--config", `model_reasoning_effort="high"`,
		"--config", "sandbox_workspace_write.network_access=true",
		"--config", `web_search="live"`,
		"--config", `approval_policy="on-request"`,
		"--image", "/x.png",
	}
	if !reflect.DeepEqual(argv, want) {
		t.Errorf("\ngot:  %v\nwant: %v", argv, want)
	}
}

func TestWireParity_ResumeThread_ImageAfterResume(t *testing.T) {
	c := spyCodex(t, CodexOptions{})
	th := c.ResumeThread("thread-7", ThreadOptions{})
	argv, _, _ := runSpy(t, c, th, PartsInput{
		{Type: InputText, Text: "continue"},
		{Type: InputLocalImage, Path: "/p.png"},
	}, TurnOptions{})

	want := []string{
		"exec", "--experimental-json",
		"resume", "thread-7",
		"--image", "/p.png",
	}
	if !reflect.DeepEqual(argv, want) {
		t.Errorf("\ngot:  %v\nwant: %v", argv, want)
	}
}
