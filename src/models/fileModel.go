package models

import "gorm.io/gorm"

// FileRecord is the row shape shared by every file type. Each type keeps its
// own table (images, docs, audio) — the repository selects which one — so the
// columns and JSON stay exactly what they were when images and docs each had
// their own model.
type FileRecord struct {
	gorm.Model

	FileName string `json:"file_name"`
	// Folder is the file's location relative to its type's upload directory,
	// using "/" as the separator. Empty means the root, which is where every
	// file uploaded before folders existed lives.
	Folder   string `json:"folder"`
	Checksum []byte `json:"checksum"`
}

// FileType describes one kind of file the CDN stores. Everything that differs
// between images, docs and audio lives here; adding a fourth type is one entry
// in FileTypes.
type FileType struct {
	// Name is both the URL segment (/api/cdn/upload/images) and the directory
	// below uploads/.
	Name string
	// FormField is the multipart field the upload arrives in, which differs
	// from Name for historical reasons ("image", not "images").
	FormField        string
	AllowedMimeTypes map[string]bool
}

// FileTypes is the registry of supported types, keyed by URL segment.
var FileTypes = map[string]FileType{
	"images": {
		Name:      "images",
		FormField: "image",
		AllowedMimeTypes: map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
			"image/bmp":  true,
		},
	},
	"docs": {
		Name:      "docs",
		FormField: "doc",
		AllowedMimeTypes: map[string]bool{
			"text/plain":                true,
			"text/plain; charset=utf-8": true,
			"application/msword":        true,
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
			"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
			"application/pdf":       true,
			"application/rtf":       true,
			"application/x-freearc": true,
			"application/zip":       true,
		},
	},
	"audio": {
		Name:      "audio",
		FormField: "audio",
		// These are what the sniffer actually returns, which is not always
		// what the extension suggests: WAV comes back as audio/wave and OGG
		// as application/ogg.
		AllowedMimeTypes: map[string]bool{
			"audio/mpeg":      true,
			"audio/wave":      true,
			"audio/wav":       true,
			"audio/x-wav":     true,
			"audio/aiff":      true,
			"application/ogg": true,
			"audio/ogg":       true,
			"audio/flac":      true,
			"audio/midi":      true,
			// m4a and some aac files sniff as generic MP4 rather than an
			// audio type.
			"video/mp4": true,
			"audio/mp4": true,
		},
	},
}

type FileRepository interface {
	GetAll() []FileRecord
	GetByCheckSum(checksum []byte) FileRecord
	Add(file FileRecord) (string, error)
	Delete(folder, fileName string) (string, bool)
	Rename(folder, oldFileName, newFileName string) error
	// DeleteFolder removes every row in the folder and its subfolders,
	// returning how many were removed.
	DeleteFolder(folder string) int64
}
