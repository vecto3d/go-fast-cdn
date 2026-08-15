package database

import (
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
	"gorm.io/gorm"
)

const (
	DbFolder = "db_data"
	DbName   = "main.db"
)

var DB *gorm.DB

func ConnectToDB() {
	dbPath := fmt.Sprintf("%v/%s", util.ExPath, DbFolder)

	_, err := os.Stat(fmt.Sprintf("%v/%s", dbPath, DbName))
	if err != nil {
		os.Mkdir(dbPath, 0o755)
		log.Printf("DB not found, creating at %v/main.db...", dbPath)
		// The handle has to be closed, or the file stays locked for the rest
		// of the process's life on Windows.
		if file, err := os.Create(fmt.Sprintf("%v/main.db", dbPath)); err == nil {
			file.Close()
		}
	}

	database, err := gorm.Open(sqlite.Open(fmt.Sprintf("%v/main.db", dbPath)), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("Failed to connect to database!")
	}
	log.Println("Connected to database!")

	database.AutoMigrate(&models.Config{})
	migrateFileTables(database)
	DB = database
	log.Println("Database initialized!")
}
