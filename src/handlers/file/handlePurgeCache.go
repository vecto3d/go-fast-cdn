package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

// HandlePurgeCache clears the whole CDN cache at the edge. Uploads, renames and
// deletes purge the URLs they touch on their own; this is for the files that
// went stale before purging was configured, and it is admin-only because it
// costs everyone else a cache miss.
func (h *FileHandler) HandlePurgeCache(c *gin.Context) {
	if !util.PurgeConfigured() {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "Cache purging is not configured. Set CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID.",
		})
		return
	}

	if err := util.PurgeEverything(); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cache purged"})
}
