package cli

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/J4NN0/mycel/internal/llm"
)

var imageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".bmp":  true,
}

// extractImages loads any image file the message points at, so that "what is in ./shot.png?" just
// works. It also returns the paths that looked like images but could not be read, so the caller can
// say so: a silently dropped attachment looks exactly like a model that ignored the picture.
func extractImages(text string) (images []llm.Image, failed []string) {
	for _, field := range splitFields(text) {
		candidates := imagePaths(field)
		if len(candidates) == 0 {
			continue
		}

		data := readFirst(candidates)
		if data == nil {
			// Only complain about something that was clearly meant as a path. A bare "photo.jpg" in a
			// sentence is far more likely to be prose than an attachment the user expected to land.
			if strings.ContainsRune(candidates[0], filepath.Separator) {
				failed = append(failed, candidates[0])
			}
			continue
		}

		images = append(images, llm.Image{Data: data})
	}

	return images, failed
}

// readFirst returns the contents of the first candidate that can be read.
func readFirst(candidates []string) []byte {
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			return data
		}
	}
	return nil
}

// splitFields splits a message into words, keeping quoted runs and backslash-escaped spaces together.
// Dragging a file into the terminal pastes its path with the spaces escaped, which plain word
// splitting would tear in half.
func splitFields(text string) []string {
	var (
		fields  []string
		token   strings.Builder
		quote   rune
		escaped bool
	)

	flush := func() {
		if token.Len() > 0 {
			fields = append(fields, token.String())
			token.Reset()
		}
	}

	for _, r := range text {
		switch {
		case escaped:
			token.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
				break
			}
			token.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
		case unicode.IsSpace(r):
			flush()
		default:
			token.WriteRune(r)
		}
	}
	flush()

	return fields
}

// imagePaths returns the file paths a word might name, most literal first, or nothing when the word
// does not look like an image at all. There is more than one candidate because terminals and editors
// like to glue a marker onto a pasted path - "[Image #1]/Users/me/shot.png" - and the path inside is
// still the file the user meant.
func imagePaths(field string) []string {
	// Trailing punctuation belongs to the sentence, not to the file name.
	path := strings.TrimRight(field, `.,;:!?`)

	if !imageExtensions[strings.ToLower(filepath.Ext(path))] {
		return nil
	}

	if after, ok := strings.CutPrefix(path, "~/"); ok {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		return []string{filepath.Join(home, after)}
	}

	candidates := []string{path}
	if i := strings.Index(path, "/"); i > 0 {
		candidates = append(candidates, path[i:])
	}

	return candidates
}
