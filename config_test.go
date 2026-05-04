package codex

import (
	"math"
	"testing"
)

func TestTomlValue_Primitives(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"hello", `"hello"`},
		{`he said "hi"`, `"he said \"hi\""`},
		{int64(42), "42"},
		{42, "42"}, // also int
		{1.5, "1.5"},
		{true, "true"},
		{false, "false"},
	}
	for _, c := range cases {
		got, err := tomlValue(c.in, "x")
		if err != nil {
			t.Errorf("tomlValue(%v) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("tomlValue(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTomlValue_Array(t *testing.T) {
	got, err := tomlValue([]any{"a", "b", 1}, "arr")
	if err != nil {
		t.Fatal(err)
	}
	want := `["a", "b", 1]`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestTomlValue_InlineTable(t *testing.T) {
	got, err := tomlValue(map[string]any{"foo": "bar", "n": 1}, "obj")
	if err != nil {
		t.Fatal(err)
	}
	// map iteration order varies — accept both orderings.
	want1 := `{foo = "bar", n = 1}`
	want2 := `{n = 1, foo = "bar"}`
	if got != want1 && got != want2 {
		t.Errorf("got %q, want %q or %q", got, want1, want2)
	}
}

func TestTomlValue_RejectsNil(t *testing.T) {
	if _, err := tomlValue(nil, "x"); err == nil {
		t.Error("expected error for nil")
	}
}

func TestTomlValue_RejectsNonFinite(t *testing.T) {
	cases := []float64{math.Inf(1), math.Inf(-1), math.NaN()}
	for _, c := range cases {
		if _, err := tomlValue(c, "x"); err == nil {
			t.Errorf("expected error for %v", c)
		}
	}
}

func TestFormatTomlKey(t *testing.T) {
	cases := map[string]string{
		"foo":     "foo",
		"foo_bar": "foo_bar",
		"foo-bar": "foo-bar",
		"FOO123":  "FOO123",
		"with space": `"with space"`,
		"":         `""`,
		"foo.bar":  `"foo.bar"`,
	}
	for in, want := range cases {
		if got := formatTomlKey(in); got != want {
			t.Errorf("formatTomlKey(%q) = %q, want %q", in, got, want)
		}
	}
}
