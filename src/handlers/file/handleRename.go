package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
	"github.com/kevinanielsen/go-fast-cdn/src/validations"
)

func (h *FileHandler) HandleRename(c *gin.Context) {
	fileType, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	oldName := c.PostForm("filename")
	newName := c.PostForm("newname")
	folder := util.SanitizeFolder(c.PostForm("folder"))

	if err := validations.ValidateRenameInput(oldName, newName); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	filteredNewName, err := util.FilterFilename(newName)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if err := util.RenameFile(folder, oldName, filteredNewName, fileType.Name); err != nil {
		c.String(http.StatusInternalServerError, "Failed to rename file: %s", err.Error())
		return
	}

	if err := repo.Rename(folder, oldName, filteredNewName); err != nil {
		c.String(http.StatusInternalServerError, "Failed to rename file: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "File renamed successfully"})
}
