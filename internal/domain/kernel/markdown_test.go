package kernel_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
)

func TestStripMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes headers",
			input: "# Header 1\n## Header 2\n### Header 3\nSome text",
			want:  "Some text",
		},
		{
			name:  "removes bold text",
			input: "This is **bold** and this is __also bold__",
			want:  "This is bold and this is also bold",
		},
		{
			name:  "removes italic text",
			input: "This is *italic* and this is _also italic_",
			want:  "This is italic and this is also italic",
		},
		{
			name:  "removes links",
			input: "Check out [my website](https://example.com) for more",
			want:  "Check out my website for more",
		},
		{
			name:  "removes images",
			input: "Here's an image: ![alt text](image.jpg) in the text",
			want:  "Here's an image:  in the text",
		},
		{
			name:  "removes code blocks",
			input: "Here's code:\n```go\nfunc main() {}\n```\nAnd more text",
			want:  "Here's code:\n\nAnd more text",
		},
		{
			name:  "removes inline code",
			input: "Use `fmt.Println()` to print",
			want:  "Use  to print",
		},
		{
			name:  "handles combined markdown",
			input: "# Title\n\nThis is **bold** with [link](url) and `code`.\n\n```\nblock\n```",
			want:  "This is bold with link and .",
		},
		{
			name:  "handles empty string",
			input: "",
			want:  "",
		},
		{
			name:  "handles plain text",
			input: "Just plain text with no markdown",
			want:  "Just plain text with no markdown",
		},
		{
			name:  "handles nested emphasis",
			input: "This is ***bold and italic*** text",
			want:  "This is bold and italic text",
		},
		{
			name:  "preserves whitespace between words",
			input: "Word1  Word2   Word3",
			want:  "Word1  Word2   Word3",
		},
		{
			name:  "handles links with special characters",
			input: "[Click here!](https://example.com?param=value&other=123)",
			want:  "Click here!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kernel.StripMarkdown(tt.input)

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
