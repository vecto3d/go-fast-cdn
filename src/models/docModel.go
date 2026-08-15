package models

import (
	"gorm.io/gorm"
)

type Doc struct {
	gorm.Model

	FileName string `json:"file_name"`
	// Folder is the file's location relative to uploads/docs, using "/" as the
	// separator. Empty means the root, which is where every file uploaded
	// before folders existed lives.
	Folder   string `json:"folder"`
	Checksum []byte `json:"checksum"`
}

type DocRepository interface {
	GetAllDocs() []Doc
	GetDocByCheckSum(checksum []byte) Doc
	AddDoc(doc Doc) (string, error)
	DeleteDoc(folder, fileName string) (string, bool)
	RenameDoc(folder, oldFileName, newFileName string) error
}
