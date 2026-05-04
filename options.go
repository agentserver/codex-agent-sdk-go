package codex

// SandboxMode mirrors TS `SandboxMode`: codex --sandbox argument.
type SandboxMode string

const (
	SandboxReadOnly         SandboxMode = "read-only"
	SandboxWorkspaceWrite   SandboxMode = "workspace-write"
	SandboxDangerFullAccess SandboxMode = "danger-full-access"
)

// ApprovalMode mirrors TS `ApprovalMode`: codex approval_policy config.
type ApprovalMode string

const (
	ApprovalNever     ApprovalMode = "never"
	ApprovalOnRequest ApprovalMode = "on-request"
	ApprovalOnFailure ApprovalMode = "on-failure"
	ApprovalUntrusted ApprovalMode = "untrusted"
)

// ReasoningEffort mirrors TS `ModelReasoningEffort`.
type ReasoningEffort string

const (
	ReasoningMinimal ReasoningEffort = "minimal"
	ReasoningLow     ReasoningEffort = "low"
	ReasoningMedium  ReasoningEffort = "medium"
	ReasoningHigh    ReasoningEffort = "high"
	ReasoningXHigh   ReasoningEffort = "xhigh"
)

// WebSearchMode mirrors TS `WebSearchMode`.
type WebSearchMode string

const (
	WebSearchDisabled WebSearchMode = "disabled"
	WebSearchCached   WebSearchMode = "cached"
	WebSearchLive     WebSearchMode = "live"
)

// ThreadOptions mirrors TS `ThreadOptions`. See spec §"Public API" for
// the full TS-to-Go field mapping.
type ThreadOptions struct {
	Model                string
	SandboxMode          SandboxMode
	WorkingDirectory     string
	AdditionalDirs       []string
	SkipGitRepoCheck     bool
	ModelReasoningEffort ReasoningEffort
	NetworkAccessEnabled *bool
	WebSearchMode        WebSearchMode
	WebSearchEnabled     *bool
	ApprovalPolicy       ApprovalMode
}

// TurnOptions mirrors TS `TurnOptions`. The TS `signal: AbortSignal` field
// is replaced by the `ctx context.Context` parameter on Run / RunStreamed.
type TurnOptions struct {
	// OutputSchema is an arbitrary JSON-serializable Go value (typically a
	// map[string]any holding a JSON Schema). When non-nil, the SDK marshals
	// it to a temp file and passes --output-schema. The temp file is
	// cleaned up when the turn ends.
	OutputSchema any
}
