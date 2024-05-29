package main

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Occurrence represents a scheduled event
type Occurrence struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Month       int       `json:"month"`
	Day         int       `json:"day"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Notify      bool      `json:"notify"`
	Notified    bool      `json:"notified"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

var db *gorm.DB

const (
	dataDir = "data"
	dbFile  = "occurrences.db"
)

func initDB() {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.Mkdir(dataDir, os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create directory:", err)
		}
	}

	var err error
	db, err = gorm.Open(sqlite.Open(path.Join(dataDir, dbFile)), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.AutoMigrate(&Occurrence{})
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}
}

func main() {
	loadEnv()
	initDB()
	ParseTemplates()

	go CheckOccurrences()

	router := gin.Default()
	router.POST("/occurrences", addOccurrence)
	router.GET("/occurrences", getOccurrences)
	router.DELETE("/occurrences/:id", deleteOccurrence)
	router.GET("/", ShowIndexPage)

	router.Run(":3000")
}
