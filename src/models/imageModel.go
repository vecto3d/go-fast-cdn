package models

import "gorm.io/gorm"

type Image struct {
	gorm.Model

	FileName string `json:"file_name"`
	// Folder is the file's location relative to uploads/images, using "/" as
	// the separator. Empty means the root, which is where every file uploaded
	// before folders existed lives.
	Folder   string `json:"folder"`
	Checksum []byte `json:"checksum"`
}

type ImageRepository interface {
	GetAllImages() []Image
	GetImageByCheckSum(checksum []byte) Image
	AddImage(image Image) (string, error)
	DeleteImage(folder, fileName string) (string, bool)
	RenameImage(folder, oldFileName, newFileName string) error
}
