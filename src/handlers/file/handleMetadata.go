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
	// Decoders registered for their side effect, so DecodeConfig recognises
	// these formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
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
	// only images pay for it — and only best-effort: SVG is markup and AVIF and
	// HEIC have no decoder here, so those report everything except a size
	// rather than failing the request.
	if fileType.Name == "images" {
		if width, height, err := imageDimensions(filePath); err == nil {
			body["width"] = width
			body["height"] = height
		} else {
			log.Printf("No dimensions for %s: %s\n", fileName, err.Error())
		}
	}

	c.JSON(http.StatusOK, body)
}

// imageDimensions decodes just enough of an image to read its size.
func imageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
