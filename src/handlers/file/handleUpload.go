package handlers

import (
	"crypto/sha256"
	"io"
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
	allowDuplicates := c.PostForm("allow_duplicates") == "true"

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

	sniffed := fileBuffer[:read]

	// Hash the whole file, not just the sniffed prefix. Hashing 512 bytes meant
	// any two files sharing a header — every icon in a set exported by the same
	// tool, for instance — looked identical and the second was rejected as a
	// duplicate it was not.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file: %s", err.Error())
		return
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file: %s", err.Error())
		return
	}
	checksum := hasher.Sum(nil)

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

	// Scoped to the folder: the same asset genuinely belongs in more than one
	// folder, and rejecting it there was never what "already exists" should mean.
	// The caller can switch the check off when it has files that are legitimately
	// identical under different names.
	if !allowDuplicates {
		if existing := repo.GetByCheckSum(folder, checksum); len(existing.Checksum) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "File already exists in this folder: " + existing.FileName,
			})
			return
		}
	}

	destination := filepath.Join(util.ExPath, "uploads", fileType.Name, folder)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create folder: %s", err.Error())
		return
	}

	// Settle the name before the row is written, so a name already taken in
	// this folder gets a suffix instead of overwriting the file that is there
	// and leaving two rows pointing at one file.
	filteredFilename = util.UniqueName(destination, filteredFilename)

	savedFilename, err := repo.Add(models.FileRecord{
		FileName: filteredFilename,
		Folder:   folder,
		Checksum: checksum,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = c.SaveUploadedFile(fileHeader, filepath.Join(destination, savedFilename))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: %s", err.Error())
		return
	}

	// The URL may have been served before (a delete and re-upload under the
	// same name), so drop whatever the edge is holding for it.
	util.PurgeURLs([]string{purgeURL(c, fileType.Name, folder, savedFilename)})

	c.JSON(http.StatusOK, gin.H{
		"file_url": downloadURL(c, fileType.Name, folder, savedFilename),
	})
}
