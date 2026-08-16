package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

// listedFile is a stored file plus its size on disk. The size is not a column:
// it is read while listing so it cannot drift from the file, and so files that
// predate this need no backfill. Statting a folder's worth of files costs a
// few milliseconds.
type listedFile struct {
	models.FileRecord
	Size int64 `json:"size"`
}

func (h *FileHandler) HandleAll(c *gin.Context) {
	fileType, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	records := repo.GetAll()
	listed := make([]listedFile, 0, len(records))

	for _, record := range records {
		entry := listedFile{FileRecord: record}

		path := filepath.Join(util.ExPath, "uploads", fileType.Name, record.Folder, record.FileName)
		if info, err := os.Stat(path); err == nil {
			entry.Size = info.Size()
		}
		// A row whose file is missing reports zero rather than failing the
		// whole listing.

		listed = append(listed, entry)
	}

	c.JSON(http.StatusOK, listed)
}
