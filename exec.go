package codex

import "strings"

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
	for _, d := range in.ThreadOpts.AdditionalDirs {
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
