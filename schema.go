package codex

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
)

// prepareOutputSchema mirrors TS `createOutputSchemaFile`
// (outputSchemaFile.ts:6-37). Writes the schema to a fresh tempdir under
// os.TempDir() and returns (path, cleanup, error). cleanup is always
// safe to call (no-op if schema was nil) and removes the entire tempdir.
func prepareOutputSchema(schema any) (path string, cleanup func(), err error) {
	noop := func() {}
	if schema == nil {
		return "", noop, nil
	}
	if !isJSONObject(schema) {
		return "", noop, errors.New("OutputSchema must be a JSON object (map or struct)")
	}
	dir, err := os.MkdirTemp("", "codex-output-schema-")
	if err != nil {
		return "", noop, err
	}
	cleanup = func() { _ = os.RemoveAll(dir) }
	path = filepath.Join(dir, "schema.json")
	data, mErr := json.Marshal(schema)
	if mErr != nil {
		cleanup()
		return "", noop, mErr
	}
	if wErr := os.WriteFile(path, data, 0o600); wErr != nil {
		cleanup()
		return "", noop, wErr
	}
	return path, cleanup, nil
}

// isJSONObject mirrors TS `isJsonObject` (outputSchemaFile.ts:39-41).
// Accepts maps and structs; rejects scalars, slices, arrays, nil.
func isJSONObject(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map, reflect.Struct:
		return true
	case reflect.Ptr:
		if rv.IsNil() {
			return false
		}
		return isJSONObject(rv.Elem().Interface())
	default:
		return false
	}
}
