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

func (h *ImageHandler) HandleImageUpload(c *gin.Context) {
	newName := c.PostForm("filename")
	folder := util.SanitizeFolder(c.PostForm("folder"))

	fileHeader, err := c.FormFile("image")
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
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/bmp":  true,
	}

	if !allowedMimeTypes[fileType] {
		c.String(http.StatusBadRequest, "Invalid file type")
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

	image := models.Image{
		FileName: filteredFilename,
		Folder:   folder,
		Checksum: fileHashBuffer[:],
	}

	imageInDatabase := h.repo.GetImageByCheckSum(fileHashBuffer[:])
	if len(imageInDatabase.Checksum) > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "File already exists",
		})
		return
	}

	savedFilename, err := h.repo.AddImage(image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	destination := filepath.Join(util.ExPath, "uploads", "images", folder)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create folder: %s", err.Error())
		return
	}

	err = c.SaveUploadedFile(fileHeader, filepath.Join(destination, savedFilename))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: %s", err.Error())
		return
	}

	body := gin.H{
		"file_url": c.Request.Host + "/api/cdn/download/images/" + util.URLPath(folder, savedFilename),
	}

	c.JSON(http.StatusOK, body)
}
