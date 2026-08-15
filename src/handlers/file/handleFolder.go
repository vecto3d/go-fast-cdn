package handlers

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

// HandleFolderList returns every folder that exists on disk for this file type,
// including ones with no files in them yet.
func (h *FileHandler) HandleFolderList(c *gin.Context) {
	fileType, _, ok := h.resolve(c)
	if !ok {
		return
	}

	root := filepath.Join(util.ExPath, "uploads", fileType.Name)
	folders := []string{}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			// A missing root just means nothing has been uploaded yet.
			if os.IsNotExist(err) {
				return fs.SkipAll
			}
			return err
		}
		if !entry.IsDir() || path == root {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		folders = append(folders, filepath.ToSlash(relative))

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list folders",
		})
		return
	}

	sort.Strings(folders)

	c.JSON(http.StatusOK, folders)
}

// HandleFolderCreate creates an empty folder for this file type.
func (h *FileHandler) HandleFolderCreate(c *gin.Context) {
	fileType, _, ok := h.resolve(c)
	if !ok {
		return
	}

	body := struct {
		Folder string `json:"folder"`
	}{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder := util.SanitizeFolder(body.Folder)
	if folder == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Folder name is required",
		})
		return
	}

	path := filepath.Join(util.ExPath, "uploads", fileType.Name, folder)
	if _, err := os.Stat(path); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Folder already exists",
		})
		return
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create folder: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Folder created successfully",
		"folder":  folder,
	})
}

// HandleFolderDelete removes a folder, everything nested below it and the
// matching database rows. The caller is expected to have confirmed: the client
// can size the damage up front with the file list.
func (h *FileHandler) HandleFolderDelete(c *gin.Context) {
	fileType, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	folder := util.SanitizeFolder(c.Query("folder"))
	if folder == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Folder name is required",
		})
		return
	}

	path := filepath.Join(util.ExPath, "uploads", fileType.Name, folder)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Folder does not exist",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete folder",
			})
		}
		return
	}

	if err := os.RemoveAll(path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete folder: " + err.Error(),
		})
		return
	}

	deleted := repo.DeleteFolder(folder)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Folder deleted successfully",
		"folder":        folder,
		"files_deleted": deleted,
	})
}
