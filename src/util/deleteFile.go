package util

import (
	"os"
	"path/filepath"
)

// DeleteFile removes a file from the given folder below uploads/<fileType>.
// An empty folder means the file type's root directory.
func DeleteFile(folder, deletedFileName string, fileType string) error {
	filePath := filepath.Join(ExPath, "uploads", fileType, SanitizeFolder(folder), deletedFileName)

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}
