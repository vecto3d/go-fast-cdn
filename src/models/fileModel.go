package models

import "gorm.io/gorm"

// FileRecord is the row shape shared by every file type. Each type keeps its
// own table (images, docs, audio, video) — the repository selects which one —
// so the columns and JSON stay exactly what they were when images and docs each
// had their own model.
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
// between the types lives here; adding another is one entry in FileTypes.
type FileType struct {
	// Name is both the URL segment (/api/cdn/upload/images) and the directory
	// below uploads/.
	Name string
	// FormField is the multipart field the upload arrives in, which differs
	// from Name for historical reasons ("image", not "images").
	FormField string
	// Extensions is the allowlist, lowercased and including the dot. Content is
	// additionally checked against util's signature table for the formats that
	// have a reliable marker.
	Extensions map[string]bool
}

func extensions(list ...string) map[string]bool {
	set := make(map[string]bool, len(list))
	for _, extension := range list {
		set[extension] = true
	}

	return set
}

// FileTypes is the registry of supported types, keyed by URL segment.
var FileTypes = map[string]FileType{
	"images": {
		Name:      "images",
		FormField: "image",
		Extensions: extensions(
			".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp",
			".svg", ".avif", ".heic", ".heif", ".tif", ".tiff", ".ico",
		),
	},
	"docs": {
		Name:      "docs",
		FormField: "doc",
		Extensions: extensions(
			".txt", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
			".pdf", ".rtf", ".arc", ".zip",
		),
	},
	"audio": {
		Name:      "audio",
		FormField: "audio",
		Extensions: extensions(
			".mp3", ".wav", ".ogg", ".oga", ".flac", ".m4a", ".aac", ".aiff", ".mid", ".midi",
		),
	},
	"video": {
		Name:      "video",
		FormField: "video",
		Extensions: extensions(
			".mp4", ".webm", ".mkv", ".mov", ".m4v", ".avi",
		),
	},
}

type FileRepository interface {
	GetAll() []FileRecord
	GetByCheckSum(folder string, checksum []byte) FileRecord
	Add(file FileRecord) (string, error)
	Delete(folder, fileName string) (string, bool)
	Rename(folder, oldFileName, newFileName string) error
	// DeleteFolder removes every row in the folder and its subfolders,
	// returning how many were removed.
	DeleteFolder(folder string) int64
}
