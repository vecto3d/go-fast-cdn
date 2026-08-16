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

func (h *FileHandler) HandleUpload(c *gin.Context) {
	fileType, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	newName := c.PostForm("filename")
	folder := util.SanitizeFolder(c.PostForm("folder"))

	fileHeader, err := c.FormFile(fileType.FormField)
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

	read, err := file.Read(fileBuffer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file: %s", err.Error())
		return
	}

	// The checksum keeps hashing the whole padded buffer, so that the values
	// stored for files uploaded before this check still match.
	fileHashBuffer := md5.Sum(fileBuffer)
	sniffed := fileBuffer[:read]

	filename := fileHeader.Filename
	if newName != "" {
		filename = newName + filepath.Ext(fileHeader.Filename)
	}

	// The extension has to be one this type accepts, and the content has to
	// back it up wherever the format carries a recognisable marker.
	extension := util.Extension(filename)
	if !fileType.Extensions[extension] {
		c.String(http.StatusBadRequest, "Invalid file type for %s: %s", fileType.Name, extension)
		return
	}
	if !util.ContentMatchesExtension(filename, sniffed) {
		c.String(http.StatusBadRequest, "File content does not match its %s extension", extension)
		return
	}

	filteredFilename, err := util.FilterFilename(filename)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if existing := repo.GetByCheckSum(fileHashBuffer[:]); len(existing.Checksum) > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "File already exists",
		})
		return
	}

	savedFilename, err := repo.Add(models.FileRecord{
		FileName: filteredFilename,
		Folder:   folder,
		Checksum: fileHashBuffer[:],
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	destination := filepath.Join(util.ExPath, "uploads", fileType.Name, folder)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create folder: %s", err.Error())
		return
	}

	err = c.SaveUploadedFile(fileHeader, filepath.Join(destination, savedFilename))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_url": downloadURL(c, fileType.Name, folder, savedFilename),
	})
}
