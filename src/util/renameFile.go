package util

import (
	"os"
	"path/filepath"
)

// RenameFile renames a file within the given folder below uploads/<fileType>.
// Renaming never moves a file between folders; an empty folder means the file
// type's root directory.
func RenameFile(folder, oldName, newName, fileType string) error {
	prefix := filepath.Join(ExPath, "uploads", fileType, SanitizeFolder(folder))

	err := os.Rename(
		filepath.Join(prefix, oldName),
		filepath.Join(prefix, newName),
	)
	if err != nil {
		return err
	}

	return nil
}
