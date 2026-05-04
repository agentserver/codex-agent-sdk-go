package codex

import "testing"

func TestSandboxModeConstants(t *testing.T) {
	cases := map[SandboxMode]string{
		SandboxReadOnly:         "read-only",
		SandboxWorkspaceWrite:   "workspace-write",
		SandboxDangerFullAccess: "danger-full-access",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestApprovalModeConstants(t *testing.T) {
	cases := map[ApprovalMode]string{
		ApprovalNever:     "never",
		ApprovalOnRequest: "on-request",
		ApprovalOnFailure: "on-failure",
		ApprovalUntrusted: "untrusted",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestReasoningEffortConstants(t *testing.T) {
	cases := map[ReasoningEffort]string{
		ReasoningMinimal: "minimal",
		ReasoningLow:     "low",
		ReasoningMedium:  "medium",
		ReasoningHigh:    "high",
		ReasoningXHigh:   "xhigh",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestWebSearchModeConstants(t *testing.T) {
	cases := map[WebSearchMode]string{
		WebSearchDisabled: "disabled",
		WebSearchCached:   "cached",
		WebSearchLive:     "live",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestThreadOptionsZeroValue(t *testing.T) {
	var o ThreadOptions
	if o.NetworkAccessEnabled != nil { t.Error("NetworkAccessEnabled should be nil zero") }
	if o.WebSearchEnabled != nil     { t.Error("WebSearchEnabled should be nil zero") }
	if o.SkipGitRepoCheck            { t.Error("SkipGitRepoCheck should be false zero") }
	if o.AdditionalDirs != nil       { t.Error("AdditionalDirs should be nil zero") }
}
