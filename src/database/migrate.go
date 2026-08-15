package database

import (
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"gorm.io/gorm"
)

// Migrate runs database migrations for all model structs using
// the global DB instance. This would typically be called on app startup.
func Migrate() {
	DB.AutoMigrate(&models.User{}, &models.UserSession{}, &models.PasswordReset{})
	migrateFileTables(DB)
}

// migrateFileTables gives every file type its own table with the shared
// FileRecord shape, so adding a type creates its table on the next start.
func migrateFileTables(db *gorm.DB) {
	for name := range models.FileTypes {
		db.Table(TableFor(name)).AutoMigrate(&models.FileRecord{})
	}
}
