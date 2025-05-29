package post_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/post"
)

func TestNewPostContent(t *testing.T) {
	t.Run("creates post content with valid input", func(t *testing.T) {
		// Create content without trailing spaces to avoid trim issues
		content300 := strings.Repeat("a", 300)                                                      // exactly 300 chars
		content10000 := strings.Repeat("a", 10000)                                                  // exactly 10000 chars
		contentWords := "word" + strings.Repeat(" word", 74)                                        // 375 chars total
		contentSentences := "Hello world!" + strings.Repeat(" Hello world!", 24)                    // 325 chars total
		contentMixed := "This is a valid post content." + strings.Repeat(" More content here.", 15) // 316 chars total

		validContents := []string{
			content300,
			content10000,
			contentWords,
			contentSentences,
			contentMixed,
		}

		for i, content := range validContents {
			t.Run(fmt.Sprintf("valid_content_%d", i), func(t *testing.T) {
				got, err := post.NewPostContent(content)

				assertNoError(t, err)
				if got.String() != content {
					t.Errorf("content mismatch at index %d", i)
				}
			})
		}
	})

	t.Run("trims whitespace from content", func(t *testing.T) {
		// Create content without trailing spaces
		coreContent := "This is valid content." + strings.Repeat(" This is valid content.", 14) // 360 chars
		input := "   " + coreContent + "   "

		got, err := post.NewPostContent(input)

		assertNoError(t, err)
		if got.String() != coreContent {
			t.Error("whitespace trimming failed")
		}
	})

	t.Run("accepts content at minimum length", func(t *testing.T) {
		content := strings.Repeat("a", post.MinPostContentLength)

		got, err := post.NewPostContent(content)

		assertNoError(t, err)
		if len(got.String()) != post.MinPostContentLength {
			t.Errorf("expected length %d, got %d", post.MinPostContentLength, len(got.String()))
		}
	})

	t.Run("accepts content at maximum length", func(t *testing.T) {
		content := strings.Repeat("a", post.MaxPostContentLength)

		got, err := post.NewPostContent(content)

		assertNoError(t, err)
		if len(got.String()) != post.MaxPostContentLength {
			t.Errorf("expected length %d, got %d", post.MaxPostContentLength, len(got.String()))
		}
	})

	t.Run("rejects empty content", func(t *testing.T) {
		_, err := post.NewPostContent("")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects whitespace only content", func(t *testing.T) {
		_, err := post.NewPostContent("   \n\t   ")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects content below minimum length", func(t *testing.T) {
		shortContent := strings.Repeat("a", post.MinPostContentLength-1)

		_, err := post.NewPostContent(shortContent)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects content above maximum length", func(t *testing.T) {
		longContent := strings.Repeat("a", post.MaxPostContentLength+1)

		_, err := post.NewPostContent(longContent)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("handles unicode content correctly", func(t *testing.T) {
		// Build content piece by piece to ensure exact length
		unicodeContent := "This is a test with unicode characters: café, naïve, résumé." +
			strings.Repeat(" More unicode: 北京, 東京, München.", 8) +
			" Additional content to meet minimum length requirements."

		got, err := post.NewPostContent(unicodeContent)

		assertNoError(t, err)
		if got.String() != unicodeContent {
			t.Error("unicode content not preserved correctly")
		}
	})

	t.Run("handles markdown content", func(t *testing.T) {
		markdownContent := `# This is a title

This is a paragraph with **bold** and *italic* text.

## Subheading

Here's a list:
- Item 1
- Item 2
- Item 3

And some code:
` + "```go\nfunc main() {\n    fmt.Println(\"Hello World\")\n}\n```" + `

More content to reach minimum length. More content to reach minimum length.
More content to reach minimum length. More content to reach minimum length.`

		got, err := post.NewPostContent(markdownContent)

		assertNoError(t, err)
		if got.String() != markdownContent {
			t.Error("markdown content not preserved correctly")
		}
	})

	t.Run("handles HTML content", func(t *testing.T) {
		htmlContent := `<p>This is HTML content with <strong>bold</strong> and <em>italic</em> text.</p>
<ul>
<li>List item 1</li>
<li>List item 2</li>
</ul>` + strings.Repeat("<p>More HTML content here.</p>", 8)

		got, err := post.NewPostContent(htmlContent)

		assertNoError(t, err)
		if got.String() != htmlContent {
			t.Error("HTML content not preserved correctly")
		}
	})
}

func TestPostContent_String(t *testing.T) {
	want := strings.Repeat("Test content ", 25) // 325 chars
	content := post.PostContent(want)

	got := content.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPostContent_Validate(t *testing.T) {
	t.Run("valid content passes", func(t *testing.T) {
		content := post.PostContent(strings.Repeat("Valid content. ", 25)) // 375 chars

		err := content.Validate()

		assertNoError(t, err)
	})

	t.Run("empty content fails", func(t *testing.T) {
		content := post.PostContent("")

		err := content.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("whitespace only content fails", func(t *testing.T) {
		content := post.PostContent("   \n\t   ")

		err := content.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("content below minimum length fails", func(t *testing.T) {
		content := post.PostContent(strings.Repeat("a", post.MinPostContentLength-1))

		err := content.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("content above maximum length fails", func(t *testing.T) {
		content := post.PostContent(strings.Repeat("a", post.MaxPostContentLength+1))

		err := content.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("content at minimum length passes", func(t *testing.T) {
		content := post.PostContent(strings.Repeat("a", post.MinPostContentLength))

		err := content.Validate()

		assertNoError(t, err)
	})

	t.Run("content at maximum length passes", func(t *testing.T) {
		content := post.PostContent(strings.Repeat("a", post.MaxPostContentLength))

		err := content.Validate()

		assertNoError(t, err)
	})

	t.Run("content just above minimum length passes", func(t *testing.T) {
		content := post.PostContent(strings.Repeat("a", post.MinPostContentLength+1))

		err := content.Validate()

		assertNoError(t, err)
	})

	t.Run("content just below maximum length passes", func(t *testing.T) {
		content := post.PostContent(strings.Repeat("a", post.MaxPostContentLength-1))

		err := content.Validate()

		assertNoError(t, err)
	})

	t.Run("validates unicode content correctly", func(t *testing.T) {
		// Create content with enough unicode characters to meet minimum length
		unicodeContent := strings.Repeat("café naïve résumé ", 20) // ~18 chars * 20 = 360 chars

		content := post.PostContent(unicodeContent)

		err := content.Validate()

		assertNoError(t, err)
	})

	t.Run("validates mixed content types", func(t *testing.T) {
		mixedContent := "Text with numbers 123, symbols !@#$%^&*(), and unicode café. " +
			strings.Repeat("More mixed content here. ", 10) // Should be well > 300 chars

		content := post.PostContent(mixedContent)

		err := content.Validate()

		assertNoError(t, err)
	})
}

func TestPostContentConstants(t *testing.T) {
	t.Run("minimum content length constant", func(t *testing.T) {
		if post.MinPostContentLength != 300 {
			t.Errorf("MinPostContentLength: got %d, want %d", post.MinPostContentLength, 300)
		}
	})

	t.Run("maximum content length constant", func(t *testing.T) {
		if post.MaxPostContentLength != 10000 {
			t.Errorf("MaxPostContentLength: got %d, want %d", post.MaxPostContentLength, 10000)
		}
	})

	t.Run("constants are logical", func(t *testing.T) {
		if post.MinPostContentLength >= post.MaxPostContentLength {
			t.Error("MinPostContentLength should be less than MaxPostContentLength")
		}
	})
}

func TestPostContent_EdgeCases(t *testing.T) {
	t.Run("content with only newlines and spaces", func(t *testing.T) {
		content := strings.Repeat("\n ", post.MinPostContentLength)

		_, err := post.NewPostContent(content)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("content with tabs and spaces", func(t *testing.T) {
		content := strings.Repeat("\t ", post.MinPostContentLength)

		_, err := post.NewPostContent(content)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("very long single word", func(t *testing.T) {
		content := strings.Repeat("a", post.MinPostContentLength)

		got, err := post.NewPostContent(content)

		assertNoError(t, err)
		if len(got.String()) != post.MinPostContentLength {
			t.Error("single long word not handled correctly")
		}
	})

	t.Run("content with repeated punctuation", func(t *testing.T) {
		// Build content without trailing spaces
		content := "Hello!" + strings.Repeat(" Hello!", 49) // 350 chars total (7*50)

		got, err := post.NewPostContent(content)

		assertNoError(t, err)
		if strings.Count(got.String(), "!") != 50 {
			t.Error("punctuation not preserved correctly")
		}
	})

	t.Run("content with mixed line endings", func(t *testing.T) {
		content := "Line 1\nLine 2\r\nLine 3\r" +
			strings.Repeat(" More content here.", 15) // Ensure well above min length

		got, err := post.NewPostContent(content)

		assertNoError(t, err)
		if !strings.Contains(got.String(), "Line 1") {
			t.Error("line endings not handled correctly")
		}
	})
}

func TestPostContent_TypeBehavior(t *testing.T) {
	t.Run("PostContent is a string type", func(t *testing.T) {
		content := post.PostContent("test")
		str := string(content)

		if str != "test" {
			t.Error("PostContent should be convertible to string")
		}
	})

	t.Run("PostContent can be compared", func(t *testing.T) {
		content1 := post.PostContent("same content")
		content2 := post.PostContent("same content")
		content3 := post.PostContent("different content")

		if content1 != content2 {
			t.Error("identical PostContent should be equal")
		}

		if content1 == content3 {
			t.Error("different PostContent should not be equal")
		}
	})

	t.Run("PostContent preserves exact input", func(t *testing.T) {
		original := "Exact content with spaces   and\ttabs\nand newlines" +
			strings.Repeat(" and more.", 30) // Ensure well above minimum

		content, err := post.NewPostContent(original)

		assertNoError(t, err)
		if content.String() != strings.TrimSpace(original) {
			t.Error("content should preserve structure after trimming")
		}
	})
}
