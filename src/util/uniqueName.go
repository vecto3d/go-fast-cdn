package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// maxNameAttempts bounds the search for a free name. Reaching it means
// thousands of files share one name, at which point the original is returned
// and the upload simply overwrites, as it did before.
const maxNameAttempts = 1000

// UniqueName returns fileName if nothing in dir is called that, and otherwise
// the same name with a numeric suffix: photo.png, photo-2.png, photo-3.png.
//
// Without this, uploading a different file under a name already in the folder
// overwrote the stored file while adding a second row, so the listing showed
// two entries pointing at one file and the original was gone.
func UniqueName(dir, fileName string) string {
	if _, err := os.Stat(filepath.Join(dir, fileName)); os.IsNotExist(err) {
		return fileName
	}

	extension := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, extension)

	for attempt := 2; attempt < maxNameAttempts; attempt++ {
		candidate := fmt.Sprintf("%s-%d%s", base, attempt, extension)
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}

	return fileName
}
