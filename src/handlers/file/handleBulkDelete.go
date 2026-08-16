package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

// maxBulkDelete bounds one request so a single call cannot tie the server up
// indefinitely. The client splits anything larger into batches.
const maxBulkDelete = 500

// HandleBulkDelete removes many files from one folder in a single request.
// Deleting a folder's worth of files one request at a time meant hundreds of
// parallel calls, which is both slow and enough concurrent traffic to trigger a
// token refresh storm on the client.
func (h *FileHandler) HandleBulkDelete(c *gin.Context) {
	fileType, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	body := struct {
		Folder string   `json:"folder"`
		Files  []string `json:"files"`
	}{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(body.Files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files given"})
		return
	}
	if len(body.Files) > maxBulkDelete {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Too many files in one request",
			"max":   maxBulkDelete,
		})
		return
	}

	folder := util.SanitizeFolder(body.Folder)

	deleted := []string{}
	failed := []gin.H{}

	for _, fileName := range body.Files {
		name, found := repo.Delete(folder, fileName)
		if !found {
			failed = append(failed, gin.H{"file": fileName, "error": "not found"})
			continue
		}

		// The row is already gone, so a file that cannot be removed from disk
		// is reported rather than retried: leaving the row behind would make
		// the listing lie about what exists.
		if err := util.DeleteFile(folder, name, fileType.Name); err != nil {
			failed = append(failed, gin.H{"file": fileName, "error": err.Error()})
			continue
		}

		deleted = append(deleted, name)
	}

	purge := make([]string, 0, len(deleted))
	for _, name := range deleted {
		purge = append(purge, purgeURL(c, fileType.Name, folder, name))
	}
	util.PurgeURLs(purge)

	c.JSON(http.StatusOK, gin.H{
		"deleted":       len(deleted),
		"failed":        failed,
		"deleted_files": deleted,
	})
}
