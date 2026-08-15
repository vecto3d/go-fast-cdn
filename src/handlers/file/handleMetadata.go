package handlers

import (
	"errors"
	"image"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
	_ "golang.org/x/image/webp"
)

func (h *FileHandler) HandleMetadata(c *gin.Context) {
	fileType, _, ok := h.resolve(c)
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
	filePath := filepath.Join(util.ExPath, "uploads", fileType.Name, folder, fileName)

	fileinfo, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File does not exist",
			})
		} else {
			log.Printf("Failed to get the file %s: %s\n", fileName, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	body := gin.H{
		"filename":     fileName,
		"folder":       folder,
		"download_url": downloadURL(c, fileType.Name, folder, fileName),
		"file_size":    fileinfo.Size(),
	}

	// Dimensions are the one piece of metadata that needs the file decoded, so
	// only images pay for it.
	if fileType.Name == "images" {
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Failed to open the image %s: %s\n", fileName, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			log.Printf("Failed to decode image %s: %s\n", fileName, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			return
		}

		body["width"] = img.Bounds().Dx()
		body["height"] = img.Bounds().Dy()
	}

	c.JSON(http.StatusOK, body)
}
