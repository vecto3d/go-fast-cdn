// Package handlers serves every file type through one set of handlers.
// Images, docs and audio differ only in their directory, multipart field and
// allowed MIME types, all of which come from models.FileTypes.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/database"
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

type FileHandler struct {
	repos map[string]models.FileRepository
}

func NewFileHandler() *FileHandler {
	repos := make(map[string]models.FileRepository, len(models.FileTypes))
	for name := range models.FileTypes {
		repos[name] = database.NewFileRepo(database.DB, name)
	}

	return &FileHandler{repos: repos}
}

// Repo exposes a type's repository for handlers that need one directly, such
// as the dashboard.
func (h *FileHandler) Repo(fileType string) models.FileRepository {
	return h.repos[fileType]
}

// resolve reads the :type parameter and returns its configuration. An unknown
// type is rejected here, before any path is built from it, so the parameter
// can never select a directory of its own.
func (h *FileHandler) resolve(c *gin.Context) (models.FileType, models.FileRepository, bool) {
	name := c.Param("type")

	fileType, ok := models.FileTypes[name]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unknown file type: " + name,
		})
		return models.FileType{}, nil, false
	}

	return fileType, h.repos[name], true
}

// downloadURL builds the public URL for a stored file, escaping each segment so
// names containing spaces still resolve.
func downloadURL(c *gin.Context, fileType, folder, fileName string) string {
	return c.Request.Host + "/api/cdn/download/" + fileType + "/" + util.URLPath(folder, fileName)
}

// purgeURL is the same URL with a scheme, which is the form Cloudflare purges
// by. Empty when there is no host to build one from.
func purgeURL(c *gin.Context, fileType, folder, fileName string) string {
	base := util.PublicBaseURL(c.Request.Host)
	if base == "" {
		return ""
	}

	return base + "/api/cdn/download/" + fileType + "/" + util.URLPath(folder, fileName)
}
