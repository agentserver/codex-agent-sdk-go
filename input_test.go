package codex

import (
	"reflect"
	"testing"
)

func TestJoinTextParts_String(t *testing.T) {
	prompt, images := joinTextParts(StringInput("hello"))
	if prompt != "hello" {
		t.Errorf("prompt = %q, want %q", prompt, "hello")
	}
	if images != nil {
		t.Errorf("images = %v, want nil", images)
	}
}

func TestJoinTextParts_Parts_TextOnly(t *testing.T) {
	prompt, images := joinTextParts(PartsInput{
		{Type: InputText, Text: "first"},
		{Type: InputText, Text: "second"},
	})
	if prompt != "first\n\nsecond" {
		t.Errorf("prompt = %q", prompt)
	}
	if images != nil {
		t.Errorf("images = %v", images)
	}
}

func TestJoinTextParts_Parts_Mixed(t *testing.T) {
	prompt, images := joinTextParts(PartsInput{
		{Type: InputText, Text: "describe"},
		{Type: InputLocalImage, Path: "/a.png"},
		{Type: InputText, Text: "and this"},
		{Type: InputLocalImage, Path: "/b.png"},
	})
	if prompt != "describe\n\nand this" {
		t.Errorf("prompt = %q", prompt)
	}
	if !reflect.DeepEqual(images, []string{"/a.png", "/b.png"}) {
		t.Errorf("images = %v", images)
	}
}

func TestJoinTextParts_EmptyParts(t *testing.T) {
	prompt, images := joinTextParts(PartsInput{})
	if prompt != "" {
		t.Errorf("prompt = %q", prompt)
	}
	if images != nil {
		t.Errorf("images = %v", images)
	}
}

func TestInput_SealedInterface(t *testing.T) {
	// Compile-time check: only StringInput and PartsInput satisfy Input.
	var _ Input = StringInput("")
	var _ Input = PartsInput(nil)
}
