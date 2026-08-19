package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	listFilesName = "list_files"
	listFilesDesc = "List what is in a folder on the user's machine: the names of the files and folders it holds, with sizes. Give it a full path. Use it to find out what is actually there before reading anything, and to work down through folders a level at a time."

	maxListEntries = 200
)

var _ Tool = (*ListFiles)(nil)

type ListFiles struct {
	log logger.Logger
}

func NewListFiles(log logger.Logger) Tool {
	log.Debugf("Tool loaded: %s", listFilesName)

	return &ListFiles{log: log}
}

func (t *ListFiles) Info() (string, string) {
	return listFilesName, listFilesDesc
}

func (t *ListFiles) Definition() schemas.ChatTool {
	const listPathDesc = "Full path of the folder to list, such as /Users/you/notes or ~/notes. Use the path exactly as the user gave it, or exactly as another tool reported it."

	props := schemas.NewOrderedMapFromPairs(
		schemas.Pair{Key: "path", Value: map[string]string{"type": "string", "description": listPathDesc}},
	)

	return schemas.ChatTool{
		Type: schemas.ChatToolTypeFunction,
		Function: &schemas.ChatToolFunction{
			Name:        listFilesName,
			Description: schemas.Ptr(listFilesDesc),
			Parameters: &schemas.ToolFunctionParameters{
				Type:       "object",
				Properties: props,
				Required:   []string{"path"},
			},
		},
	}
}

type listFilesArgs struct {
	Path string `json:"path"`
}

func (t *ListFiles) Execute(_ context.Context, raw json.RawMessage) (string, error) {
	var a listFilesArgs
	err := json.Unmarshal(raw, &a)
	if err != nil {
		return "", fmt.Errorf("parse list_files args: %w", err)
	}

	path, err := resolvePath(a.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", pathError("list", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is a file, not a folder: read it with %s instead", path, readFileName)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", pathError("list", path, err)
	}

	t.log.Debugf("Listed %s (%d entries)", path, len(entries))

	if len(entries) == 0 {
		return fmt.Sprintf("%s is empty.", path), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s holds:\n", path)

	for i, e := range entries {
		if i == maxListEntries {
			fmt.Fprintf(&sb, "\n[first %d of %d shown: there are more]", maxListEntries, len(entries))
			break
		}
		fmt.Fprintf(&sb, "- %s\n", describeEntry(e))
	}

	return sb.String(), nil
}

func describeEntry(e os.DirEntry) string {
	switch {
	case e.IsDir():
		return e.Name() + "/"
	case e.Type()&os.ModeSymlink != 0:
		return e.Name() + " (symlink)"
	case !e.Type().IsRegular():
		return e.Name() + " (special file)"
	}

	info, err := e.Info()
	if err != nil {
		return e.Name()
	}

	return fmt.Sprintf("%s (%s)", e.Name(), humanSize(info.Size()))
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes) / unit
	for _, suffix := range []string{"KB", "MB", "GB"} {
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
		value /= unit
	}

	return fmt.Sprintf("%.1f TB", value)
}
