package codex

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
)

// tomlValue mirrors TS `toTomlValue` (exec.ts:262-296). Renders a Go value
// as a TOML literal suitable for `codex --config key=<value>`.
func tomlValue(v any, path string) (string, error) {
	switch x := v.(type) {
	case nil:
		return "", fmt.Errorf("config override at %s cannot be null", path)
	case string:
		return strconv.Quote(x), nil
	case bool:
		if x {
			return "true", nil
		}
		return "false", nil
	case int:
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return "", fmt.Errorf("config override at %s must be a finite number", path)
		}
		return strconv.FormatFloat(x, 'g', -1, 64), nil
	case []any:
		parts := make([]string, len(x))
		for i, elem := range x {
			s, err := tomlValue(elem, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return "", err
			}
			parts[i] = s
		}
		return "[" + joinComma(parts) + "]", nil
	case map[string]any:
		// Sort keys for deterministic output across map iteration order.
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(x))
		for _, k := range keys {
			child := x[k]
			if k == "" {
				return "", fmt.Errorf("config override keys must be non-empty strings")
			}
			if child == nil {
				continue
			}
			s, err := tomlValue(child, path+"."+k)
			if err != nil {
				return "", err
			}
			parts = append(parts, formatTomlKey(k)+" = "+s)
		}
		return "{" + joinComma(parts) + "}", nil
	default:
		return "", fmt.Errorf("unsupported config override value at %s: %T", path, v)
	}
}

var tomlBareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// formatTomlKey mirrors TS `formatTomlKey` (exec.ts:309-313).
func formatTomlKey(k string) string {
	if tomlBareKey.MatchString(k) {
		return k
	}
	return strconv.Quote(k)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
