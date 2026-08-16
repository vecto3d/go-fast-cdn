package handlers

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

// folderMetaFile holds a folder's display name. It lives inside the folder so
// it travels with the directory and needs no table of its own; the leading dot
// keeps it out of the way.
const folderMetaFile = ".folder.json"

type folderMeta struct {
	DisplayName string `json:"display_name"`
}

// Folder is a folder as the UI sees it: the path is what appears in URLs and on
// disk, the name is what gets displayed.
type Folder struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func metaPath(fileType, folder string) string {
	return filepath.Join(util.ExPath, "uploads", fileType, folder, folderMetaFile)
}

// displayName reads a folder's stored name, falling back to its last path
// segment for folders created before display names existed.
func displayName(fileType, folder string) string {
	fallback := folder
	if index := strings.LastIndex(folder, "/"); index >= 0 {
		fallback = folder[index+1:]
	}

	content, err := os.ReadFile(metaPath(fileType, folder))
	if err != nil {
		return fallback
	}

	meta := folderMeta{}
	if err := json.Unmarshal(content, &meta); err != nil || meta.DisplayName == "" {
		return fallback
	}

	return meta.DisplayName
}

// HandleFolderList returns every folder that exists on disk for this file type,
// including ones with no files in them yet.
func (h *FileHandler) HandleFolderList(c *gin.Context) {
	fileType, _, ok := h.resolve(c)
	if !ok {
		return
	}

	root := filepath.Join(util.ExPath, "uploads", fileType.Name)
	folders := []Folder{}

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
		slashed := filepath.ToSlash(relative)
		folders = append(folders, Folder{Path: slashed, Name: displayName(fileType.Name, slashed)})

		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list folders",
		})
		return
	}

	sort.Slice(folders, func(i, j int) bool { return folders[i].Path < folders[j].Path })

	c.JSON(http.StatusOK, folders)
}

// HandleFolderCreate creates an empty folder for this file type. The path is
// slugified so it stays URL-safe; the name the user typed is kept for display.
func (h *FileHandler) HandleFolderCreate(c *gin.Context) {
	fileType, _, ok := h.resolve(c)
	if !ok {
		return
	}

	body := struct {
		Folder      string `json:"folder"`
		DisplayName string `json:"display_name"`
	}{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder := util.SanitizeFolder(util.Slugify(body.Folder))
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

	name := strings.TrimSpace(body.DisplayName)
	if name != "" && name != folder {
		meta, _ := json.Marshal(folderMeta{DisplayName: name})
		// A folder without its metadata still works, it just displays its
		// path, so a write failure here is not worth failing the request over.
		_ = os.WriteFile(metaPath(fileType.Name, folder), meta, 0o644)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Folder created successfully",
		"folder":  folder,
		"name":    displayName(fileType.Name, folder),
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
