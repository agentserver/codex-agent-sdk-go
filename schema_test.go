package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareOutputSchema_Nil(t *testing.T) {
	path, cleanup, err := prepareOutputSchema(nil)
	if err != nil { t.Fatal(err) }
	if path != "" { t.Errorf("path = %q, want empty", path) }
	if cleanup == nil { t.Error("cleanup should be non-nil even when nil schema") }
	cleanup() // should not panic
}

func TestPrepareOutputSchema_Map(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"x": map[string]any{"type": "string"},
		},
	}
	path, cleanup, err := prepareOutputSchema(schema)
	if err != nil { t.Fatal(err) }
	defer cleanup()

	if !strings.HasSuffix(path, "schema.json") {
		t.Errorf("path = %q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("temp file should exist: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil { t.Fatal(err) }
	if got["type"] != "object" { t.Errorf("got %v", got) }

	// Cleanup removes the dir
	dir := filepath.Dir(path)
	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("dir should be gone after cleanup, stat err = %v", err)
	}
}

func TestPrepareOutputSchema_RejectsNonObject(t *testing.T) {
	for _, bad := range []any{42, "string", []any{1, 2}, true} {
		_, _, err := prepareOutputSchema(bad)
		if err == nil {
			t.Errorf("expected error for %v (%T)", bad, bad)
		}
	}
}
