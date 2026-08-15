package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *FileHandler) HandleAll(c *gin.Context) {
	_, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, repo.GetAll())
}
