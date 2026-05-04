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

// CodexOptions mirrors TS `CodexOptions`. Top-level configuration for
// the Codex handle: binary location, model endpoint override, API auth,
// always-applied config overrides, and env-var policy.
type CodexOptions struct {
	// Path to the codex binary. Default "codex" (PATH lookup).
	BinaryPath string
	// Becomes `-c openai_base_url="<value>"` on every spawn (NOT an env var).
	BaseURL string
	// Set as CODEX_API_KEY env var on every spawn.
	APIKey string
	// Extra TOML-typed config overrides; flattened to dotted-key form
	// and serialized as TOML literals. Applied as `-c key=value` on every
	// spawn, BEFORE per-thread/per-turn flags.
	Config map[string]any
	// If non-nil, replaces process env entirely. If nil, current process
	// env is inherited. After composition: if
	// CODEX_INTERNAL_ORIGINATOR_OVERRIDE is unset → "codex_sdk_go";
	// if APIKey != "" → CODEX_API_KEY = APIKey.
	Env map[string]string
}
