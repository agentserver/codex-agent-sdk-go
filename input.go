package codex

import "strings"

// Input is the prompt parameter to Run / RunStreamed. Mirrors TS
// `Input = string | UserInput[]`. Use StringInput or PartsInput.
type Input interface{ codexInput() }

// StringInput is a bare-text prompt. Equivalent to
// PartsInput{{Type: InputText, Text: string(s)}}.
type StringInput string

func (StringInput) codexInput() {}

// PartsInput is a sequence of typed input parts. Mirrors TS `UserInput[]`.
type PartsInput []UserInput

func (PartsInput) codexInput() {}

// UserInput mirrors TS:
//   {type: "text", text: string} | {type: "local_image", path: string}
// Set Text when Type==InputText; set Path when Type==InputLocalImage.
type UserInput struct {
	Type UserInputType
	Text string
	Path string
}

type UserInputType string

const (
	InputText       UserInputType = "text"
	InputLocalImage UserInputType = "local_image"
)

// joinTextParts mirrors TS `normalizeInput` (thread.ts:140-156): texts are
// joined with "\n\n", images are extracted to a parallel slice for
// --image flags.
func joinTextParts(input Input) (prompt string, images []string) {
	switch v := input.(type) {
	case StringInput:
		return string(v), nil
	case PartsInput:
		var texts []string
		for _, p := range v {
			switch p.Type {
			case InputText:
				texts = append(texts, p.Text)
			case InputLocalImage:
				images = append(images, p.Path)
			}
		}
		return strings.Join(texts, "\n\n"), images
	}
	return "", nil
}
