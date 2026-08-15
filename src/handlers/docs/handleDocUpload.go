package handlers

import (
	"crypto/md5"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

func (h *DocHandler) HandleDocUpload(c *gin.Context) {
	fileHeader, err := c.FormFile("doc")
	newName := c.PostForm("filename")
	folder := util.SanitizeFolder(c.PostForm("folder"))

	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read file: %s", err.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to open file: %s", err.Error())
		return
	}
	defer file.Close()

	fileBuffer := make([]byte, 512)
	_, err = file.Read(fileBuffer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file: %s", err.Error())
		return
	}
	fileType := http.DetectContentType(fileBuffer)

	allowedMimeTypes := map[string]bool{
		"text/plain":                true,
		"text/plain; charset=utf-8": true,
		"application/msword":        true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/pdf":       true,
		"application/rtf":       true,
		"application/x-freearc": true,
		"application/zip":       true,
	}

	if !allowedMimeTypes[fileType] {
		c.String(http.StatusBadRequest, "Invalid file type: %s", fileType)
		return
	}

	fileHashBuffer := md5.Sum(fileBuffer)
	var filename string
	if newName == "" {
		filename = fileHeader.Filename
	} else {
		filename = newName + filepath.Ext(fileHeader.Filename)
	}

	filteredFilename, err := util.FilterFilename(filename)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	doc := models.Doc{
		FileName: filteredFilename,
		Folder:   folder,
		Checksum: fileHashBuffer[:],
	}

	docInDatabase := h.repo.GetDocByCheckSum(fileHashBuffer[:])
	if len(docInDatabase.Checksum) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "File already exists"})
		return
	}

	savedFileName, err := h.repo.AddDoc(doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	destination := filepath.Join(util.ExPath, "uploads", "docs", folder)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create folder: %s", err.Error())
		return
	}

	err = c.SaveUploadedFile(fileHeader, filepath.Join(destination, savedFileName))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: %s", err.Error())
		return
	}

	body := gin.H{
		"file_url": c.Request.Host + "/api/cdn/download/docs/" + util.URLPath(folder, savedFileName),
	}

	c.JSON(http.StatusOK, body)
}
