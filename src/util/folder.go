package util

import (
	"net/url"
	"strings"
)

// SanitizeFolder normalizes a user-supplied folder path into a safe relative
// path below the uploads directory. Empty, "." and ".." segments are dropped
// and the illegal characters FilterFilename rejects are stripped from each
// remaining segment, so the result can never escape its file type's directory.
// The zero value ("") means the file type's root.
func SanitizeFolder(folder string) string {
	var segments []string

	for _, segment := range strings.Split(strings.ReplaceAll(folder, `\`, "/"), "/") {
		segment = strings.TrimSpace(segment)
		if segment == "" || segment == "." || segment == ".." {
			continue
		}
		segments = append(segments, segment)
	}

	return strings.Join(segments, "/")
}

// URLPath joins path parts into a URL path, escaping each segment so that
// names containing spaces or other reserved characters still produce a URL
// that resolves. Empty parts are skipped, so a root-level file (folder "")
// yields just the escaped filename.
func URLPath(parts ...string) string {
	var segments []string

	for _, part := range parts {
		for _, segment := range strings.Split(part, "/") {
			if segment != "" {
				segments = append(segments, url.PathEscape(segment))
			}
		}
	}

	return strings.Join(segments, "/")
}
