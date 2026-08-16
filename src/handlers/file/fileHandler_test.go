package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/database"
	"github.com/kevinanielsen/go-fast-cdn/src/initializers"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
	"github.com/stretchr/testify/require"
)

// newTestHandler points the app at a scratch directory with its own database,
// so each test gets an empty CDN.
func newTestHandler(t *testing.T) *FileHandler {
	t.Helper()

	util.ExPath = t.TempDir()
	initializers.CreateFolders()
	database.ConnectToDB()
	t.Cleanup(func() {
		// The sqlite file has to be closed before the temp directory can be
		// removed, which matters on Windows where an open file cannot be
		// deleted.
		if sqlDB, err := database.DB.DB(); err == nil {
			sqlDB.Close()
		}
		os.Remove(fmt.Sprintf("%s/%s/%s", util.ExPath, database.DbFolder, database.DbName))
	})

	return NewFileHandler()
}

// uploadRequest builds a multipart upload for the given type, with the file
// arriving in the named form field.
func uploadRequest(t *testing.T, fileType, formField, fileName, folder string, content []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(formField, fileName)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.WriteField("folder", folder))
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cdn/upload/"+fileType, body)
	c.Request.Header.Add("Content-Type", writer.FormDataContentType())
	c.Params = gin.Params{{Key: "type", Value: fileType}}

	return c, w
}

func TestHandleUpload(t *testing.T) {
	jpegBytes := encodeJPEG(t, 64, 48)
	textBytes := bytes.Repeat([]byte("plain text file. "), 40)
	// A RIFF container that is a WAV, to prove "RIFF" alone is not taken as WebP.
	wavBytes := append([]byte{0x52, 0x49, 0x46, 0x46, 0, 0, 0, 0, 0x57, 0x41, 0x56, 0x45}, bytes.Repeat([]byte{0}, 600)...)

	tests := []struct {
		name       string
		fileType   string
		formField  string
		fileName   string
		folder     string
		content    []byte
		wantStatus int
	}{
		{"image at root", "images", "image", "photo.jpg", "", jpegBytes, http.StatusOK},
		{"image in folder", "images", "image", "photo.jpg", "holiday/2026", jpegBytes, http.StatusOK},
		{"doc", "docs", "doc", "notes.txt", "reports", textBytes, http.StatusOK},
		{"audio", "audio", "audio", "song.flac", "tracks", append([]byte("fLaC"), bytes.Repeat([]byte{7}, 600)...), http.StatusOK},
		{"video", "video", "video", "clip.webm", "clips", append([]byte{0x1A, 0x45, 0xDF, 0xA3}, bytes.Repeat([]byte{0}, 600)...), http.StatusOK},
		{"svg counts as an image", "images", "image", "logo.svg", "", []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`), http.StatusOK},
		{"unknown type", "videos", "video", "clip.mp4", "", jpegBytes, http.StatusBadRequest},
		{"extension not allowed for type", "audio", "audio", "photo.jpg", "", jpegBytes, http.StatusBadRequest},
		{"content contradicts extension", "images", "image", "photo.png", "", jpegBytes, http.StatusBadRequest},
		{"wav renamed as webp is caught", "images", "image", "fake.webp", "", wavBytes, http.StatusBadRequest},
		{"missing form field", "images", "wrongfield", "photo.jpg", "", jpegBytes, http.StatusBadRequest},
		{"filename with two periods", "images", "image", "photo.small.jpg", "", jpegBytes, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(t)

			c, w := uploadRequest(t, tt.fileType, tt.formField, tt.fileName, tt.folder, tt.content)

			handler.HandleUpload(c)

			require.Equal(t, tt.wantStatus, w.Result().StatusCode, w.Body.String())

			if tt.wantStatus == http.StatusOK {
				// The file landed in its folder, and the row knows where.
				require.FileExists(t, filepath.Join(util.ExPath, "uploads", tt.fileType, tt.folder, tt.fileName))

				files := handler.Repo(tt.fileType).GetAll()
				require.Len(t, files, 1)
				require.Equal(t, tt.folder, files[0].Folder)
			}
		})
	}
}

// Files that share a header but differ later are different files. Hashing only
// the sniffed prefix made every icon exported by the same tool look identical,
// so a batch upload rejected most of itself as duplicates.
func TestHandleUploadAcceptsFilesSharingAHeader(t *testing.T) {
	handler := newTestHandler(t)

	first := encodeJPEG(t, 64, 48)
	second := append(append([]byte{}, first...), bytes.Repeat([]byte{0x42}, 64)...)
	require.Equal(t, first[:512], second[:512], "the test needs two files with an identical prefix")

	c, w := uploadRequest(t, "images", "image", "icon-a.jpg", "icons", first)
	handler.HandleUpload(c)
	require.Equal(t, http.StatusOK, w.Result().StatusCode, w.Body.String())

	cc, ww := uploadRequest(t, "images", "image", "icon-b.jpg", "icons", second)
	handler.HandleUpload(cc)
	require.Equal(t, http.StatusOK, ww.Result().StatusCode, ww.Body.String())

	require.Len(t, handler.Repo("images").GetAll(), 2)
}

// The same bytes in a different folder is a legitimate copy, not a duplicate.
func TestHandleUploadAllowsSameFileInAnotherFolder(t *testing.T) {
	handler := newTestHandler(t)
	content := encodeJPEG(t, 64, 48)

	c, w := uploadRequest(t, "images", "image", "logo.jpg", "brand/light", content)
	handler.HandleUpload(c)
	require.Equal(t, http.StatusOK, w.Result().StatusCode, w.Body.String())

	cc, ww := uploadRequest(t, "images", "image", "logo.jpg", "brand/dark", content)
	handler.HandleUpload(cc)
	require.Equal(t, http.StatusOK, ww.Result().StatusCode, ww.Body.String())

	require.Len(t, handler.Repo("images").GetAll(), 2)
}

func TestHandleUploadRejectsDuplicate(t *testing.T) {
	handler := newTestHandler(t)
	content := encodeJPEG(t, 64, 48)

	c, first := uploadRequest(t, "images", "image", "photo.jpg", "gallery", content)
	handler.HandleUpload(c)
	require.Equal(t, http.StatusOK, first.Result().StatusCode)

	cc, second := uploadRequest(t, "images", "image", "photo-copy.jpg", "gallery", content)
	handler.HandleUpload(cc)

	require.Equal(t, http.StatusConflict, second.Result().StatusCode)
	require.Contains(t, second.Body.String(), "already exists in this folder")
}

// The duplicate check is refusable: a set can legitimately contain the same
// bytes under two names.
func TestHandleUploadAllowsDuplicatesWhenAsked(t *testing.T) {
	handler := newTestHandler(t)
	content := encodeJPEG(t, 64, 48)

	c, first := uploadRequest(t, "images", "image", "outfit-a.jpg", "police", content)
	handler.HandleUpload(c)
	require.Equal(t, http.StatusOK, first.Result().StatusCode)

	// Same bytes, different name, with the check switched off.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "outfit-b.jpg")
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.WriteField("folder", "police"))
	require.NoError(t, writer.WriteField("allow_duplicates", "true"))
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	cc, _ := gin.CreateTestContext(w)
	cc.Request = httptest.NewRequest(http.MethodPost, "/api/cdn/upload/images", body)
	cc.Request.Header.Add("Content-Type", writer.FormDataContentType())
	cc.Params = gin.Params{{Key: "type", Value: "images"}}

	handler.HandleUpload(cc)

	require.Equal(t, http.StatusOK, w.Result().StatusCode, w.Body.String())
	require.Len(t, handler.Repo("images").GetAll(), 2)
}

// Two different files under one name must not clobber each other.
func TestHandleUploadSuffixesATakenName(t *testing.T) {
	handler := newTestHandler(t)

	c, first := uploadRequest(t, "images", "image", "outfit.jpg", "police", encodeJPEG(t, 64, 48))
	handler.HandleUpload(c)
	require.Equal(t, http.StatusOK, first.Result().StatusCode)

	cc, second := uploadRequest(t, "images", "image", "outfit.jpg", "police", encodeJPEG(t, 32, 24))
	handler.HandleUpload(cc)
	require.Equal(t, http.StatusOK, second.Result().StatusCode, second.Body.String())

	names := []string{}
	for _, file := range handler.Repo("images").GetAll() {
		names = append(names, file.FileName)
	}
	require.ElementsMatch(t, []string{"outfit.jpg", "outfit-2.jpg"}, names)

	// Both files are on disk: the first was not overwritten.
	require.FileExists(t, filepath.Join(util.ExPath, "uploads", "images", "police", "outfit.jpg"))
	require.FileExists(t, filepath.Join(util.ExPath, "uploads", "images", "police", "outfit-2.jpg"))
}

func TestHandleMetadata(t *testing.T) {
	handler := newTestHandler(t)

	dir := filepath.Join(util.ExPath, "uploads", "images", "holiday")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "my photo.jpg"), encodeJPEG(t, 32, 16), 0o644))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/cdn/images/my%20photo.jpg?folder=holiday", nil)
	c.Params = gin.Params{{Key: "type", Value: "images"}, {Key: "filename", Value: "my photo.jpg"}}

	handler.HandleMetadata(c)

	require.Equal(t, http.StatusOK, w.Result().StatusCode, w.Body.String())

	result := map[string]any{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	require.Equal(t, "my photo.jpg", result["filename"])
	require.Equal(t, "holiday", result["folder"])
	require.Equal(t, float64(32), result["width"])
	require.Equal(t, float64(16), result["height"])
	// The space has to be escaped or the URL does not resolve.
	require.Contains(t, result["download_url"], "holiday/my%20photo.jpg")
}

func TestHandleMetadataMissingFile(t *testing.T) {
	handler := newTestHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/cdn/images/nope.jpg", nil)
	c.Params = gin.Params{{Key: "type", Value: "images"}, {Key: "filename", Value: "nope.jpg"}}

	handler.HandleMetadata(c)

	require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestFolderLifecycle(t *testing.T) {
	handler := newTestHandler(t)

	create := func(folder string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		payload, _ := json.Marshal(map[string]string{"folder": folder, "display_name": "Display " + folder})
		c.Request = httptest.NewRequest(http.MethodPost, "/api/cdn/folder/images", bytes.NewReader(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "type", Value: "images"}}
		handler.HandleFolderCreate(c)
		return w
	}

	require.Equal(t, http.StatusOK, create("logos/dark").Result().StatusCode)
	// Creating the same folder twice is a conflict, not a silent success.
	require.Equal(t, http.StatusConflict, create("logos/dark").Result().StatusCode)
	// A traversal attempt is stripped, never applied.
	require.Equal(t, http.StatusOK, create("../../escaped").Result().StatusCode)
	require.NoDirExists(t, filepath.Join(util.ExPath, "..", "..", "escaped"))
	require.DirExists(t, filepath.Join(util.ExPath, "uploads", "images", "escaped"))

	// An empty folder survives, which is the point of creating one up front.
	listed := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(listed)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/cdn/folder/images", nil)
	c.Params = gin.Params{{Key: "type", Value: "images"}}
	handler.HandleFolderList(c)

	folders := []Folder{}
	require.NoError(t, json.NewDecoder(listed.Body).Decode(&folders))
	paths := []string{}
	for _, folder := range folders {
		paths = append(paths, folder.Path)
	}
	require.Equal(t, []string{"escaped", "logos", "logos/dark"}, paths)

	// The typed name is kept for display; the path stays the slug.
	byPath := map[string]string{}
	for _, folder := range folders {
		byPath[folder.Path] = folder.Name
	}
	require.Equal(t, "Display logos/dark", byPath["logos/dark"])
	// An intermediate folder nobody named falls back to its own segment.
	require.Equal(t, "logos", byPath["logos"])
}

func TestFolderCreateSlugifiesTheName(t *testing.T) {
	handler := newTestHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	payload, _ := json.Marshal(map[string]string{"folder": "My Cats 2026!", "display_name": "My Cats 2026!"})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cdn/folder/images", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "type", Value: "images"}}

	handler.HandleFolderCreate(c)

	require.Equal(t, http.StatusOK, w.Result().StatusCode, w.Body.String())

	result := map[string]any{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	require.Equal(t, "my-cats-2026", result["folder"])
	require.Equal(t, "My Cats 2026!", result["name"])
	require.DirExists(t, filepath.Join(util.ExPath, "uploads", "images", "my-cats-2026"))
}

func TestHandleFolderDeleteIsRecursive(t *testing.T) {
	handler := newTestHandler(t)

	// A file nested two levels down, both on disk and in the database.
	c, upload := uploadRequest(t, "images", "image", "photo.jpg", "logos/dark", encodeJPEG(t, 8, 8))
	handler.HandleUpload(c)
	require.Equal(t, http.StatusOK, upload.Result().StatusCode)

	w := httptest.NewRecorder()
	cc, _ := gin.CreateTestContext(w)
	cc.Request = httptest.NewRequest(http.MethodDelete, "/api/cdn/folder/images?folder=logos", nil)
	cc.Params = gin.Params{{Key: "type", Value: "images"}}

	handler.HandleFolderDelete(cc)

	require.Equal(t, http.StatusOK, w.Result().StatusCode, w.Body.String())

	result := map[string]any{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	require.Equal(t, float64(1), result["files_deleted"])
	require.NoDirExists(t, filepath.Join(util.ExPath, "uploads", "images", "logos"))
	require.Empty(t, handler.Repo("images").GetAll())
}

func TestHandleFolderDeleteRefusesRoot(t *testing.T) {
	handler := newTestHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/cdn/folder/images?folder=", nil)
	c.Params = gin.Params{{Key: "type", Value: "images"}}

	handler.HandleFolderDelete(c)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	require.DirExists(t, filepath.Join(util.ExPath, "uploads"))
}

// Helper functions

func encodeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}

	buffer := &bytes.Buffer{}
	require.NoError(t, encodeImage(buffer, img))

	return buffer.Bytes()
}

func encodeImage(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
}
