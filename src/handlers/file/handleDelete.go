package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

func (h *FileHandler) HandleDelete(c *gin.Context) {
	fileType, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	fileName := c.Param("filename")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File name is required",
		})
		return
	}

	folder := util.SanitizeFolder(c.Query("folder"))

	deletedFileName, success := repo.Delete(folder, fileName)
	if !success {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	if err := util.DeleteFile(folder, deletedFileName, fileType.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file",
		})
		return
	}

	// Without this the file stays downloadable from the edge for the rest of
	// its cache lifetime, despite being gone from the origin.
	util.PurgeURLs([]string{purgeURL(c, fileType.Name, folder, deletedFileName)})

	c.JSON(http.StatusOK, gin.H{
		"message":  "File deleted successfully",
		"fileName": deletedFileName,
	})
}
