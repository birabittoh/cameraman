package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// Occurrence represents a scheduled event
type Occurrence struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Day          uint   `json:"day"`
	Month        uint   `json:"month"`
	Year         *uint  `json:"year"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Notify       bool   `json:"notify"`
	Notified     bool   `json:"notified"`
	NotifiedSoft bool   `json:"notified_soft"`
}

const (
	dataDir                       = "data"
	dbFile                        = "occurrences.db"
	defaultNotificationWindow     = 5
	defaultSoftNotificationWindow = 2
	defaultSleepDuration          = 1
	defaultPort                   = "3000"
)

var (
	db           *gorm.DB
	port         string
	allowedHosts []string
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

	SoftNotificationWindow, err = strconv.Atoi(os.Getenv("DAYS_BEFORE_SOFT_NOTIFICATION"))
	if err != nil {
		SoftNotificationWindow = defaultSoftNotificationWindow
	}
	log.Println("Soft notification window (days):", SoftNotificationWindow)

	loadedSleepDuration, err := strconv.Atoi(os.Getenv("HOURS_BETWEEN_CHECKS"))
	if err != nil {
		SleepDuration = defaultSleepDuration * time.Hour
	} else {
		SleepDuration = time.Duration(loadedSleepDuration) * time.Hour
	}
	log.Println("Sleep duration:", SleepDuration)

	port = os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	if hosts := os.Getenv("ALLOWED_HOSTS"); hosts != "" {
		for _, h := range strings.Split(hosts, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				allowedHosts = append(allowedHosts, h)
			}
		}
	}
	log.Println("Allowed hosts:", allowedHosts)
}

func main() {
	loadEnv()
	initDB()
	ParseTemplates()

	go CheckOccurrences()

	http.HandleFunc("GET /", ShowIndexPage)

	http.HandleFunc("GET /occurrences", getOccurrences)
	http.HandleFunc("POST /occurrences", addOccurrence)
	http.HandleFunc("DELETE /occurrences/{id}", deleteOccurrence)

	log.Println("Starting server at port " + port + "...")
	if err := http.ListenAndServe(":"+port, corsMiddleware(http.DefaultServeMux)); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, h := range allowedHosts {
			if origin == h {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				break
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
