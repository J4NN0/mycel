package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	grepFilesName = "grep_files"
	grepFilesDesc = "Search the text of the files in a folder and return the lines that match, each with its file and line number. Give it a full path and a pattern. This is how you find where something is written without opening files one by one; read_file afterwards when you need the text around a match."

	maxMatches     = 50
	maxGrepLine    = 200
	maxGrepFiles   = 2000
	maxGrepFileSze = 1 << 20
)

var skippedDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"__pycache__":  true,
}

var _ Tool = (*GrepFiles)(nil)

type GrepFiles struct {
	log logger.Logger
}

func NewGrepFiles(log logger.Logger) Tool {
	log.Debugf("Tool loaded: %s", grepFilesName)

	return &GrepFiles{log: log}
}

func (t *GrepFiles) Info() (string, string) {
	return grepFilesName, grepFilesDesc
}

func (t *GrepFiles) Definition() schemas.ChatTool {
	const (
		grepPatternDesc = "What to look for, as a regular expression. Plain words work as themselves. An all-lowercase pattern matches regardless of case; include a capital letter to match case exactly."
		grepPathDesc    = "Full path of the folder to search, such as /Users/you/notes or ~/notes. Every file below it is searched, so name the narrowest folder that could hold what you are after."
	)

	props := schemas.NewOrderedMapFromPairs(
		schemas.Pair{Key: "pattern", Value: map[string]string{"type": "string", "description": grepPatternDesc}},
		schemas.Pair{Key: "path", Value: map[string]string{"type": "string", "description": grepPathDesc}},
	)

	return schemas.ChatTool{
		Type: schemas.ChatToolTypeFunction,
		Function: &schemas.ChatToolFunction{
			Name:        grepFilesName,
			Description: schemas.Ptr(grepFilesDesc),
			Parameters: &schemas.ToolFunctionParameters{
				Type:       "object",
				Properties: props,
				Required:   []string{"pattern", "path"},
			},
		},
	}
}

type grepFilesArgs struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
}

type grepMatch struct {
	path string
	line int
	text string
}

func (t *GrepFiles) Execute(ctx context.Context, raw json.RawMessage) (string, error) {
	var a grepFilesArgs
	err := json.Unmarshal(raw, &a)
	if err != nil {
		return "", fmt.Errorf("parse grep_files args: %w", err)
	}

	expression, err := compilePattern(a.Pattern)
	if err != nil {
		return "", err
	}

	root, err := resolvePath(a.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(root)
	if err != nil {
		return "", pathError("search", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is a file, not a folder: read it with %s instead", root, readFileName)
	}

	matches, scanned, stopped, err := search(ctx, root, expression)
	if err != nil {
		return "", err
	}

	t.log.Debugf("Searched %d file(s) under %s for %q: %d match(es)", scanned, root, a.Pattern, len(matches))

	if len(matches) == 0 {
		return fmt.Sprintf("Nothing in %s matches %q (%d files searched).", root, a.Pattern, scanned), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d matching line(s) for %q in %s:\n\n", len(matches), a.Pattern, root)
	for _, m := range matches {
		fmt.Fprintf(&sb, "%s:%d: %s\n", m.path, m.line, m.text)
	}
	if stopped {
		fmt.Fprintf(&sb, "\n[stopped early: narrow the pattern or search a smaller folder to see the rest]")
	}

	return sb.String(), nil
}

// search walks the tree without following symlinks, so a link cannot send it round in a circle. It
// gives up after a bounded amount of work: the folder it is pointed at may be enormous.
func search(ctx context.Context, root string, expression *regexp.Regexp) (matches []grepMatch, scanned int, stopped bool, err error) {
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if entry.IsDir() {
			if path != root && skippedDirs[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		if info, statErr := entry.Info(); statErr == nil && info.Size() > maxGrepFileSze {
			return nil
		}

		if scanned == maxGrepFiles {
			stopped = true
			return filepath.SkipAll
		}
		scanned++

		found, full := searchFile(path, expression, len(matches))
		matches = append(matches, found...)
		if full {
			stopped = true
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return nil, scanned, stopped, fmt.Errorf("search %s: %w", root, err)
	}

	return matches, scanned, stopped, nil
}

func searchFile(path string, expression *regexp.Regexp, found int) ([]grepMatch, bool) {
	data, err := os.ReadFile(path)
	if err != nil || isBinary(data) {
		return nil, false
	}

	var matches []grepMatch
	for i, line := range strings.Split(string(data), "\n") {
		if !expression.MatchString(line) {
			continue
		}
		if found+len(matches) == maxMatches {
			return matches, true
		}
		matches = append(matches, grepMatch{path: path, line: i + 1, text: clip(line, maxGrepLine)})
	}

	return matches, false
}

// compilePattern treats an all-lowercase pattern as case-insensitive, the way ripgrep and friends do.
func compilePattern(pattern string) (*regexp.Regexp, error) {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return nil, fmt.Errorf("no pattern given: say what you are looking for")
	}

	if trimmed == strings.ToLower(trimmed) {
		trimmed = "(?i)" + trimmed
	}

	expression, err := regexp.Compile(trimmed)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid search pattern: %w", pattern, err)
	}

	return expression, nil
}
