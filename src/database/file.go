package database

import (
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"gorm.io/gorm"
)

// fileRepo is one repository over any of the file tables; the table name is
// what makes it an image, doc or audio repository.
type fileRepo struct {
	DB    *gorm.DB
	table string
}

// NewFileRepo returns a repository for the given file type name (the same
// names as models.FileTypes: images, docs, audio).
func NewFileRepo(db *gorm.DB, fileType string) models.FileRepository {
	return &fileRepo{DB: db, table: TableFor(fileType)}
}

// TableFor maps a file type name to its table. The tables images and docs
// predate folders and audio, so their names are kept as they were.
func TableFor(fileType string) string {
	return fileType
}

func (repo *fileRepo) query() *gorm.DB {
	return repo.DB.Table(repo.table).Model(&models.FileRecord{})
}

func (repo *fileRepo) GetAll() []models.FileRecord {
	var entries []models.FileRecord

	repo.query().Find(&entries)

	return entries
}

func (repo *fileRepo) GetByCheckSum(checksum []byte) models.FileRecord {
	var entry models.FileRecord

	repo.query().Where("checksum = ?", checksum).First(&entry)

	return entry
}

func (repo *fileRepo) Add(file models.FileRecord) (string, error) {
	result := repo.query().Create(&file)
	if result.Error != nil {
		return "", result.Error
	}

	return file.FileName, nil
}

func (repo *fileRepo) Delete(folder, fileName string) (string, bool) {
	var entry models.FileRecord

	result := repo.query().Where("folder = ? AND file_name = ?", folder, fileName).First(&entry)
	if result.Error != nil {
		return "", false
	}

	repo.query().Delete(&entry)

	return fileName, true
}

func (repo *fileRepo) Rename(folder, oldFileName, newFileName string) error {
	return repo.query().
		Where("folder = ? AND file_name = ?", folder, oldFileName).
		Update("file_name", newFileName).Error
}

func (repo *fileRepo) DeleteFolder(folder string) int64 {
	// The folder itself plus everything nested below it.
	result := repo.query().
		Where("folder = ? OR folder LIKE ?", folder, folder+"/%").
		Delete(&models.FileRecord{})

	return result.RowsAffected
}
