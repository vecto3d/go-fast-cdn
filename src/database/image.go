package database

import (
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"gorm.io/gorm"
)

type imageRepo struct {
	DB *gorm.DB
}

func NewImageRepo(db *gorm.DB) models.ImageRepository {
	return &imageRepo{DB: db}
}

func (repo *imageRepo) GetAllImages() []models.Image {
	var entries []models.Image

	repo.DB.Find(&entries, &models.Image{})

	return entries
}

func (repo *imageRepo) GetImageByCheckSum(checksum []byte) models.Image {
	var entries models.Image

	repo.DB.Where("checksum = ?", checksum).First(&entries)

	return entries
}

func (repo *imageRepo) AddImage(image models.Image) (string, error) {
	result := repo.DB.Create(&image)
	if result.Error != nil {
		return "", result.Error
	}

	return image.FileName, nil
}

func (repo *imageRepo) DeleteImage(folder, fileName string) (string, bool) {
	var image models.Image

	result := repo.DB.Where("folder = ? AND file_name = ?", folder, fileName).First(&image)

	if result.Error == nil {
		repo.DB.Delete(&image)
		return fileName, true
	} else {
		return "", false
	}
}

func (repo *imageRepo) RenameImage(folder, oldFileName, newFileName string) error {
	image := models.Image{}
	return repo.DB.Model(&image).Where("folder = ? AND file_name = ?", folder, oldFileName).Update("file_name", newFileName).Error
}
