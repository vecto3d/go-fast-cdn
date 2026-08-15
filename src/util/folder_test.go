package util

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeFolder(t *testing.T) {
	tests := []struct {
		name   string
		folder string
		want   string
	}{
		{"empty stays root", "", ""},
		{"simple folder", "logos", "logos"},
		{"nested folder", "logos/dark", "logos/dark"},
		{"leading and trailing slashes", "/logos/dark/", "logos/dark"},
		{"repeated slashes", "logos//dark", "logos/dark"},
		{"surrounding whitespace", " logos / dark ", "logos/dark"},
		{"backslashes are separators", `logos\dark`, "logos/dark"},
		{"dot segments dropped", "./logos/./dark", "logos/dark"},
		{"parent segments dropped", "../logos/../../dark", "logos/dark"},
		{"traversal only", "../../..", ""},
		{"absolute traversal", "/../../etc", "etc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeFolder(tt.folder); got != tt.want {
				t.Errorf("SanitizeFolder(%q) = %q, want %q", tt.folder, got, tt.want)
			}
		})
	}
}

// The whole point of SanitizeFolder is that the path it returns cannot reach
// outside the upload directory, so assert that on the joined result too.
func TestSanitizeFolderCannotEscapeUploads(t *testing.T) {
	uploads := filepath.Join("/srv", "uploads", "images")

	for _, folder := range []string{"../../etc", `..\..\etc`, "logos/../../..", "/etc/passwd"} {
		joined := filepath.Join(uploads, SanitizeFolder(folder))
		if !strings.HasPrefix(joined, uploads) {
			t.Errorf("folder %q escaped the uploads directory: %q", folder, joined)
		}
	}
}

func TestURLPath(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{"root level file", []string{"", "cat.png"}, "cat.png"},
		{"spaces are encoded", []string{"", "my cat.png"}, "my%20cat.png"},
		{"folder and file", []string{"logos/dark", "logo.png"}, "logos/dark/logo.png"},
		{"spaces in folder", []string{"my logos", "my cat.png"}, "my%20logos/my%20cat.png"},
		{"reserved characters", []string{"", "a+b&c?.png"}, "a+b&c%3F.png"},
		{"separators survive escaping", []string{"a/b", "c.png"}, "a/b/c.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := URLPath(tt.parts...); got != tt.want {
				t.Errorf("URLPath(%q) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}
