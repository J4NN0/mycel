package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/J4NN0/mycel/internal/logger"
)

// testTree lays out a small folder to read, list and search.
func testTree(t *testing.T) string {
	t.Helper()

	base := t.TempDir()
	for _, dir := range []string{filepath.Join(base, "projects"), filepath.Join(base, ".git"), filepath.Join(base, "empty")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	write(filepath.Join(base, "readme.md"), "# Title\nthe needle is here\nlast line\n")
	write(filepath.Join(base, "projects", "alpha.md"), "nothing to see\nNEEDLE in caps\n")
	write(filepath.Join(base, ".git", "config"), "the needle in git\n")
	write(filepath.Join(base, "logo.png"), "PNG\x00\x01binary")

	return base
}

func run(t *testing.T, tl Tool, args any) (string, error) {
	t.Helper()

	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	return tl.Execute(context.Background(), raw)
}

func testLog() logger.Logger { return logger.New("test", "error") }

func TestReadFileTool(t *testing.T) {
	base := testTree(t)
	tl := NewReadFile(testLog())

	out, err := run(t, tl, readFileArgs{Path: filepath.Join(base, "readme.md")})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "the needle is here") {
		t.Errorf("contents missing:\n%s", out)
	}
	if !strings.Contains(out, filepath.Join(base, "readme.md")) {
		t.Errorf("the path read is not reported back:\n%s", out)
	}
}

func TestReadFileToolRejections(t *testing.T) {
	base := testTree(t)
	tl := NewReadFile(testLog())

	cases := map[string]string{
		"empty path":   "  ",
		"missing file": filepath.Join(base, "nope.md"),
		"a folder":     base,
		"binary file":  filepath.Join(base, "logo.png"),
	}
	for name, path := range cases {
		if _, err := run(t, tl, readFileArgs{Path: path}); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

func TestReadFileToolTruncates(t *testing.T) {
	base := testTree(t)
	long := filepath.Join(base, "long.txt")
	if err := os.WriteFile(long, []byte(strings.Repeat("a", readMaxChars+500)), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	out, err := run(t, NewReadFile(testLog()), readFileArgs{Path: long})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "characters only") {
		t.Errorf("a long file was cut without saying so:\n%s", out[len(out)-120:])
	}
}

func TestListFilesTool(t *testing.T) {
	base := testTree(t)
	tl := NewListFiles(testLog())

	out, err := run(t, tl, listFilesArgs{Path: base})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "projects/") {
		t.Errorf("folders are not marked as such:\n%s", out)
	}
	if !strings.Contains(out, "readme.md (") {
		t.Errorf("files are not listed with a size:\n%s", out)
	}

	out, err = run(t, tl, listFilesArgs{Path: filepath.Join(base, "empty")})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "empty") {
		t.Errorf("an empty folder was not reported as empty: %q", out)
	}

	if _, err := run(t, tl, listFilesArgs{Path: filepath.Join(base, "readme.md")}); err == nil {
		t.Error("a file was listed as a folder")
	}
}

func TestGrepFilesTool(t *testing.T) {
	base := testTree(t)
	tl := NewGrepFiles(testLog())

	out, err := run(t, tl, grepFilesArgs{Pattern: "needle", Path: base})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "readme.md:2:") {
		t.Errorf("match is missing its file and line:\n%s", out)
	}
	// A lowercase pattern is case-insensitive, so NEEDLE counts too.
	if !strings.Contains(out, "alpha.md:2:") {
		t.Errorf("case-insensitive match missed:\n%s", out)
	}
	if strings.Contains(out, ".git") {
		t.Errorf(".git was searched:\n%s", out)
	}
}

func TestGrepFilesToolCaseSensitiveWithCapitals(t *testing.T) {
	base := testTree(t)

	out, err := run(t, NewGrepFiles(testLog()), grepFilesArgs{Pattern: "NEEDLE", Path: base})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(out, "readme.md") {
		t.Errorf("a capitalised pattern matched lowercase text:\n%s", out)
	}
	if !strings.Contains(out, "alpha.md") {
		t.Errorf("the exact match was not found:\n%s", out)
	}
}

func TestGrepFilesToolNoMatches(t *testing.T) {
	base := testTree(t)

	out, err := run(t, NewGrepFiles(testLog()), grepFilesArgs{Pattern: "haystack", Path: base})
	if err != nil {
		t.Fatalf("finding nothing is not an error: %v", err)
	}
	if !strings.Contains(out, "Nothing in") {
		t.Errorf("no-match output was %q", out)
	}
}

func TestGrepFilesToolRejectsBadInput(t *testing.T) {
	base := testTree(t)
	tl := NewGrepFiles(testLog())

	if _, err := run(t, tl, grepFilesArgs{Pattern: "[unclosed", Path: base}); err == nil {
		t.Error("an invalid regexp was accepted")
	}
	if _, err := run(t, tl, grepFilesArgs{Pattern: "  ", Path: base}); err == nil {
		t.Error("an empty pattern was accepted")
	}
	if _, err := run(t, tl, grepFilesArgs{Pattern: "x", Path: filepath.Join(base, "readme.md")}); err == nil {
		t.Error("a file was accepted as a folder to search")
	}
}

// A cancelled turn must stop the walk rather than grinding through a huge tree.
func TestGrepFilesToolHonoursCancellation(t *testing.T) {
	base := testTree(t)

	raw, err := json.Marshal(grepFilesArgs{Pattern: "needle", Path: base})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := NewGrepFiles(testLog()).Execute(ctx, raw); err == nil {
		t.Error("a cancelled search reported success")
	}
}

func TestResolvePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}

	got, err := resolvePath("~/notes/todo.md")
	if err != nil {
		t.Fatalf("resolvePath: %v", err)
	}
	if want := filepath.Join(home, "notes", "todo.md"); got != want {
		t.Errorf("resolvePath(~) = %q, want %q", got, want)
	}

	if got, err := resolvePath("/tmp/../etc/hosts"); err != nil || got != "/etc/hosts" {
		t.Errorf("resolvePath did not clean the path: %q, %v", got, err)
	}
	if _, err := resolvePath(""); err == nil {
		t.Error("an empty path was accepted")
	}
}

func TestHumanSize(t *testing.T) {
	for size, want := range map[int64]string{0: "0 B", 512: "512 B", 2048: "2.0 KB", 5 << 20: "5.0 MB"} {
		if got := humanSize(size); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", size, got, want)
		}
	}
}
