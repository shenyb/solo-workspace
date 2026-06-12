package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveBlockStripsAdjacentNewlines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".bashrc")

	initial := "export PATH=$PATH\n# sw completion bash start\nsource sw\n# sw completion bash end\nalias ll='ls -la'\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := removeBlock(path, "# sw completion bash start", "# sw completion bash end"); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "export PATH=$PATH\nalias ll='ls -la'\n"
	if string(got) != want {
		t.Fatalf("after removeBlock:\n%q\nwant:\n%q", got, want)
	}
}

func TestRemoveBlockRepeatedUninstallNoBlankLineGrowth(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".bashrc")
	base := "line\n"

	for i := 0; i < 3; i++ {
		content := base + "# sw completion bash start\nsource sw\n# sw completion bash end\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := removeBlock(path, "# sw completion bash start", "# sw completion bash end"); err != nil {
			t.Fatal(err)
		}
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != base {
		t.Fatalf("repeated uninstall should leave %q, got %q", base, got)
	}
	if strings.Count(string(got), "\n\n") > 0 {
		t.Fatalf("unexpected blank line growth: %q", got)
	}
}
