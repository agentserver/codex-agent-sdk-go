package codex

import (
	"reflect"
	"testing"
)

func TestBuildArgs_Minimal(t *testing.T) {
	got, err := buildArgs(buildArgsInput{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"exec", "--experimental-json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestBuildArgs_BaseURL(t *testing.T) {
	got, err := buildArgs(buildArgsInput{
		CodexOpts: CodexOptions{BaseURL: "https://x.example/v1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"exec", "--experimental-json",
		"--config", `openai_base_url="https://x.example/v1"`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestBuildArgs_ConfigFlattenedFirst(t *testing.T) {
	got, err := buildArgs(buildArgsInput{
		CodexOpts: CodexOptions{Config: map[string]any{"a": "b"}},
		ThreadOpts: ThreadOptions{Model: "o3"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"exec", "--experimental-json",
		"--config", `a="b"`,
		"--model", "o3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestBuildArgs_AllThreadOptions(t *testing.T) {
	tt := true
	ff := false
	got, err := buildArgs(buildArgsInput{
		ThreadOpts: ThreadOptions{
			Model:                 "o3",
			SandboxMode:           SandboxWorkspaceWrite,
			WorkingDirectory:      "/tmp/w",
			AdditionalDirectories: []string{"/d1", "/d2"},
			SkipGitRepoCheck:      true,
			ModelReasoningEffort:  ReasoningHigh,
			NetworkAccessEnabled:  &tt,
			WebSearchMode:         WebSearchLive,
			ApprovalPolicy:        ApprovalOnRequest,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"exec", "--experimental-json",
		"--model", "o3",
		"--sandbox", "workspace-write",
		"--cd", "/tmp/w",
		"--add-dir", "/d1",
		"--add-dir", "/d2",
		"--skip-git-repo-check",
		"--config", `model_reasoning_effort="high"`,
		"--config", "sandbox_workspace_write.network_access=true",
		"--config", `web_search="live"`,
		"--config", `approval_policy="on-request"`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:  %v\nwant: %v", got, want)
	}
	_ = ff
}

func TestBuildArgs_LegacyWebSearchEnabled(t *testing.T) {
	tt := true
	ff := false
	gotTrue, _ := buildArgs(buildArgsInput{
		ThreadOpts: ThreadOptions{WebSearchEnabled: &tt},
	})
	gotFalse, _ := buildArgs(buildArgsInput{
		ThreadOpts: ThreadOptions{WebSearchEnabled: &ff},
	})
	wantTrue := []string{"exec", "--experimental-json", "--config", `web_search="live"`}
	wantFalse := []string{"exec", "--experimental-json", "--config", `web_search="disabled"`}
	if !reflect.DeepEqual(gotTrue, wantTrue) {
		t.Errorf("true: got %v want %v", gotTrue, wantTrue)
	}
	if !reflect.DeepEqual(gotFalse, wantFalse) {
		t.Errorf("false: got %v want %v", gotFalse, wantFalse)
	}
}

func TestBuildArgs_WebSearchModeOverridesLegacy(t *testing.T) {
	tt := true
	got, _ := buildArgs(buildArgsInput{
		ThreadOpts: ThreadOptions{
			WebSearchMode:    WebSearchCached,
			WebSearchEnabled: &tt,
		},
	})
	want := []string{"exec", "--experimental-json", "--config", `web_search="cached"`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestBuildArgs_OutputSchema(t *testing.T) {
	got, _ := buildArgs(buildArgsInput{
		OutputSchemaPath: "/tmp/schema.json",
	})
	want := []string{"exec", "--experimental-json", "--output-schema", "/tmp/schema.json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestBuildArgs_ResumeAndImagesOrdering(t *testing.T) {
	got, _ := buildArgs(buildArgsInput{
		ThreadID: "01HM-thread",
		Images:   []string{"/a.png", "/b.png"},
	})
	want := []string{
		"exec", "--experimental-json",
		"resume", "01HM-thread",
		"--image", "/a.png",
		"--image", "/b.png",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:  %v\nwant: %v", got, want)
	}
}

func TestBuildArgs_FullExample(t *testing.T) {
	tt := true
	got, err := buildArgs(buildArgsInput{
		CodexOpts: CodexOptions{
			BaseURL: "https://api.x.example/v1",
			Config:  map[string]any{"feature.x": true},
		},
		ThreadOpts: ThreadOptions{
			Model:                "o3",
			SandboxMode:          SandboxWorkspaceWrite,
			NetworkAccessEnabled: &tt,
			ApprovalPolicy:       ApprovalNever,
		},
		ThreadID:         "thread-1",
		Images:           []string{"/x.png"},
		OutputSchemaPath: "/tmp/s.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"exec", "--experimental-json",
		"--config", "feature.x=true",
		"--config", `openai_base_url="https://api.x.example/v1"`,
		"--model", "o3",
		"--sandbox", "workspace-write",
		"--output-schema", "/tmp/s.json",
		"--config", "sandbox_workspace_write.network_access=true",
		"--config", `approval_policy="never"`,
		"resume", "thread-1",
		"--image", "/x.png",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:  %v\nwant: %v", got, want)
	}
}
