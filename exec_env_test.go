package codex

import (
	"reflect"
	"testing"
)

func TestComposeEnv_NilUsesProcessEnv(t *testing.T) {
	procEnv := []string{"PATH=/usr/bin", "HOME=/h"}
	got := composeEnv(CodexOptions{}, procEnv)
	want := map[string]string{
		"PATH":                              "/usr/bin",
		"HOME":                              "/h",
		"CODEX_INTERNAL_ORIGINATOR_OVERRIDE": "codex_sdk_go",
	}
	if !reflect.DeepEqual(envToMap(got), want) {
		t.Errorf("got %v want %v", envToMap(got), want)
	}
}

func TestComposeEnv_OptOverridesProcess(t *testing.T) {
	procEnv := []string{"PATH=/usr/bin"}
	got := composeEnv(CodexOptions{
		Env: map[string]string{"FOO": "bar"},
	}, procEnv)
	want := map[string]string{
		"FOO":                                "bar",
		"CODEX_INTERNAL_ORIGINATOR_OVERRIDE": "codex_sdk_go",
	}
	if !reflect.DeepEqual(envToMap(got), want) {
		t.Errorf("got %v want %v", envToMap(got), want)
	}
}

func TestComposeEnv_PreservesUserOriginator(t *testing.T) {
	got := composeEnv(CodexOptions{
		Env: map[string]string{"CODEX_INTERNAL_ORIGINATOR_OVERRIDE": "custom"},
	}, nil)
	m := envToMap(got)
	if m["CODEX_INTERNAL_ORIGINATOR_OVERRIDE"] != "custom" {
		t.Errorf("originator = %q, want preserved", m["CODEX_INTERNAL_ORIGINATOR_OVERRIDE"])
	}
}

func TestComposeEnv_APIKeySetsCodexAPIKey(t *testing.T) {
	got := composeEnv(CodexOptions{APIKey: "sk-test"}, nil)
	m := envToMap(got)
	if m["CODEX_API_KEY"] != "sk-test" {
		t.Errorf("CODEX_API_KEY = %q", m["CODEX_API_KEY"])
	}
}

func TestComposeEnv_APIKeyOverridesProvidedEnv(t *testing.T) {
	got := composeEnv(CodexOptions{
		APIKey: "sk-new",
		Env:    map[string]string{"CODEX_API_KEY": "sk-old"},
	}, nil)
	m := envToMap(got)
	if m["CODEX_API_KEY"] != "sk-new" {
		t.Errorf("CODEX_API_KEY = %q, want sk-new", m["CODEX_API_KEY"])
	}
}

func envToMap(env []string) map[string]string {
	m := map[string]string{}
	for _, kv := range env {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				m[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	return m
}
