package kernel

import (
	"regexp"
	"strings"
)

// StripMarkdown removes basic Markdown syntax from content.
// Useful for generating plain text excerpts and accurate word counts.
func StripMarkdown(content string) string {
	// Step 1: Remove code blocks (preserve newlines)
	codeBlockRe := regexp.MustCompile("(?s)```[^`]*```")
	content = codeBlockRe.ReplaceAllStringFunc(content, func(match string) string {
		// Count newlines in the code block and replace with that many newlines
		newlineCount := strings.Count(match, "\n")
		return strings.Repeat("\n", newlineCount)
	})

	// Step 2: Remove inline code
	inlineCodeRe := regexp.MustCompile("`[^`]+`")
	content = inlineCodeRe.ReplaceAllString(content, "")

	// Step 3: Remove images
	imageRe := regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	content = imageRe.ReplaceAllString(content, "")

	// Step 4: Replace links with their text
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	content = linkRe.ReplaceAllString(content, "$1")

	// Step 5: Remove emphasis markers (bold/italic)
	// Handle from most specific to least specific
	content = regexp.MustCompile(`\*\*\*([^*]+)\*\*\*`).ReplaceAllString(content, "$1")
	content = regexp.MustCompile(`___([^_]+)___`).ReplaceAllString(content, "$1")
	content = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(content, "$1")
	content = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(content, "$1")
	content = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(content, "$1")
	content = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(content, "$1")

	// Step 6: Process headers
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Check if entire line is a header
		if regexp.MustCompile(`^\s*#{1,6}\s+`).MatchString(line) {
			// Skip header lines entirely
			continue
		}

		// Remove inline headers from remaining lines
		line = regexp.MustCompile(`#{1,6}\s+`).ReplaceAllString(line, "")

		// Keep the line (even if empty, to preserve structure)
		cleanLines = append(cleanLines, line)
	}

	// Step 7: Rejoin and clean up
	content = strings.Join(cleanLines, "\n")

	// Trim leading and trailing whitespace
	content = strings.TrimSpace(content)

	// Normalize multiple blank lines to maximum of two
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	return content
}
