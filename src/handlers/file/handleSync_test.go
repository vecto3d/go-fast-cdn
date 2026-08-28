package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevinanielsen/go-fast-cdn/src/models"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanUploadsKeysMatchRowShape(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "root.webp"))
	writeFile(t, filepath.Join(root, "clothing", "male", "jbib", "male_heist_0_0.webp"))
	writeFile(t, filepath.Join(root, "clothing", "notes.txt")) // wrong extension

	found, err := scanUploads(root, models.FileTypes["images"])
	if err != nil {
		t.Fatal(err)
	}

	// A row stores Folder relative to the type directory with "/" separators, and
	// the empty string for the root; the scan key has to line up with that or every
	// file looks new on every sync.
	want := map[string]struct{ folder, name string }{
		"/root.webp":                             {"", "root.webp"},
		"clothing/male/jbib/male_heist_0_0.webp": {"clothing/male/jbib", "male_heist_0_0.webp"},
	}

	if len(found) != len(want) {
		t.Fatalf("scanned %d files, want %d: %v", len(found), len(want), found)
	}
	for key, expected := range want {
		got, ok := found[key]
		if !ok {
			t.Fatalf("missing key %q in %v", key, found)
		}
		if got.folder != expected.folder || got.name != expected.name {
			t.Errorf("%q -> folder %q name %q, want %q / %q", key, got.folder, got.name, expected.folder, expected.name)
		}
	}
}

func TestScanUploadsMissingDirectoryIsEmptyNotAnError(t *testing.T) {
	// A type that has never had an upload has no directory. That must read as
	// "nothing stored", not as a failure, or a sync of one type breaks the others.
	found, err := scanUploads(filepath.Join(t.TempDir(), "absent"), models.FileTypes["images"])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("expected no files, got %v", found)
	}
}

func TestIsTrue(t *testing.T) {
	for _, value := range []string{"1", "true", "TRUE", "yes", "on"} {
		if !isTrue(value) {
			t.Errorf("isTrue(%q) = false", value)
		}
	}
	for _, value := range []string{"", "0", "false", "no", "maybe"} {
		if isTrue(value) {
			t.Errorf("isTrue(%q) = true", value)
		}
	}
}
