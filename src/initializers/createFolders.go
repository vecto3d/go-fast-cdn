package initializers

import (
	"fmt"
	"os"

	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

func CreateFolders() {
	uploadsFolder := fmt.Sprintf("%v/uploads", util.ExPath)
	os.Mkdir(uploadsFolder, 0o755)

	for name := range models.FileTypes {
		os.Mkdir(fmt.Sprintf("%v/%v", uploadsFolder, name), 0o755)
	}
}
