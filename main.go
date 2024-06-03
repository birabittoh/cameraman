package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Occurrence represents a scheduled event
type Occurrence struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Day         uint   `json:"day"`
	Month       uint   `json:"month"`
	Year        *uint  `json:"year"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Notify      bool   `json:"notify"`
	Notified    bool   `json:"notified"`
}

const (
	dataDir                   = "data"
	dbFile                    = "occurrences.db"
	defaultNotificationWindow = 3
	defaultSleepDuration      = 1
	defaultPort               = 3000
)

var (
	db   *gorm.DB
	port int
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
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	NotificationWindow, err = strconv.Atoi(os.Getenv("DAYS_BEFORE_NOTIFICATION"))
	if err != nil {
		NotificationWindow = defaultNotificationWindow
	}
	log.Println("Notification window (days):", NotificationWindow)

	loadedSleepDuration, err := strconv.Atoi(os.Getenv("HOURS_BETWEEN_CHECKS"))
	if err != nil {
		SleepDuration = defaultSleepDuration * time.Hour
	} else {
		SleepDuration = time.Duration(loadedSleepDuration) * time.Hour
	}
	log.Println("Sleep duration:", SleepDuration)

	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = defaultPort
	}
	log.Println("Port:", port)
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

	router.Run(fmt.Sprintf(":%d", port))
}
