package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	readFileName = "read_file"
	readFileDesc = "Read a text file on the user's machine and return what is in it. Give it a full path, the one the user gave you or one that list_files or grep_files reported. Reach for grep_files rather than opening file after file looking for something, and never guess at a path: ask the user where the file is."
	readPathDesc = "Full path of the file to read, such as /Users/you/notes/todo.md or ~/notes/todo.md. Use the path exactly as the user gave it, or exactly as another tool reported it."

	readMaxChars = 8000
	binaryProbe  = 64 << 10
)

var _ Tool = (*ReadFile)(nil)

type ReadFile struct {
	log logger.Logger
}

func NewReadFile(log logger.Logger) Tool {
	log.Debugf("Tool loaded: %s", readFileName)

	return &ReadFile{log: log}
}

func (t *ReadFile) Info() (string, string) {
	return readFileName, readFileDesc
}

func (t *ReadFile) Definition() schemas.ChatTool {
	props := schemas.NewOrderedMapFromPairs(
		schemas.Pair{Key: "path", Value: map[string]string{"type": "string", "description": readPathDesc}},
	)

	return schemas.ChatTool{
		Type: schemas.ChatToolTypeFunction,
		Function: &schemas.ChatToolFunction{
			Name:        readFileName,
			Description: schemas.Ptr(readFileDesc),
			Parameters: &schemas.ToolFunctionParameters{
				Type:       "object",
				Properties: props,
				Required:   []string{"path"},
			},
		},
	}
}

type readFileArgs struct {
	Path string `json:"path"`
}

func (t *ReadFile) Execute(_ context.Context, raw json.RawMessage) (string, error) {
	var a readFileArgs
	err := json.Unmarshal(raw, &a)
	if err != nil {
		return "", fmt.Errorf("parse read_file args: %w", err)
	}

	path, err := resolvePath(a.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", pathError("read", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a folder, not a file: list it with %s instead", path, listFilesName)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", pathError("read", path, err)
	}
	if isBinary(data) {
		return "", fmt.Errorf("%s is not a text file, so there is nothing to read out of it", path)
	}

	t.log.Debugf("Read %s (%d bytes)", path, len(data))

	text := string(data)
	truncated := utf8.RuneCountInString(text) > readMaxChars
	if truncated {
		text = string([]rune(text)[:readMaxChars])
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "File: %s\n\n%s", path, text)
	if truncated {
		fmt.Fprintf(&sb, "\n\n[…first %d characters only: the file goes on past this point]", readMaxChars)
	}

	return sb.String(), nil
}

// resolvePath expands a leading ~ and makes the path absolute, so what is reported back to the user
// is the same file the model asked for.
func resolvePath(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("no path given: ask the user which file or folder they mean")
	}

	if trimmed == "~" || strings.HasPrefix(trimmed, "~"+string(filepath.Separator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand %q: %w", trimmed, err)
		}
		trimmed = filepath.Join(home, strings.TrimPrefix(trimmed, "~"))
	}

	absolute, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("%q is not a usable path: %w", raw, err)
	}

	return filepath.Clean(absolute), nil
}

// pathError keeps the model out of a retry loop by saying which of the two likely things went wrong.
func pathError(action, path string, err error) error {
	switch {
	case os.IsNotExist(err):
		return fmt.Errorf("there is nothing at %s: check the path with the user rather than guessing another one", path)
	case os.IsPermission(err):
		return fmt.Errorf("%s cannot be read: the user's account does not have permission", path)
	default:
		return fmt.Errorf("%s %s: %w", action, path, err)
	}
}

// isBinary looks for a NUL byte, which text files do not carry and binaries almost always do early.
func isBinary(data []byte) bool {
	head := data
	if len(head) > binaryProbe {
		head = head[:binaryProbe]
	}

	return bytes.IndexByte(head, 0) != -1
}
