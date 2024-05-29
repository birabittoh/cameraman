package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

var notificationWindow = 3

func notifyTelegram(occurrence Occurrence) {
	client := resty.New()
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	threadID := os.Getenv("TELEGRAM_THREAD_ID")

	message := fmt.Sprintf("*Giorno %02d/%02d*:\n\n_%s_\n%s",
		occurrence.Day, occurrence.Month, occurrence.Name, occurrence.Description)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramToken)

	// Create the payload
	payload := map[string]interface{}{
		"message_thread_id": threadID,
		"chat_id":           chatID,
		"text":              message,
		"parse_mode":        "markdown",
	}

	// Send the POST request
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(url)

	if err != nil {
		log.Printf("Failed to send notification: %v", err)
		return
	}
	log.Printf("Notification sent: %s, Response: %s", message, resp)
}

func resetNotifications() {
	if err := db.Model(&Occurrence{}).Where("notified = ?", true).Update("notified", false).Error; err != nil {
		log.Printf("Failed to reset notifications: %v", err)
	} else {
		log.Println("Notifications have been reset for the new year.")
	}
}

func CheckOccurrences() {
	const sleepDuration = 12 * time.Hour

	for {
		now := time.Now()
		var occurrences []Occurrence
		endWindow := now.AddDate(0, 0, notificationWindow)

		db.Where("notified = ? AND ((month = ? AND day >= ?) OR (month = ? AND day <= ?))",
			false, now.Month(), now.Day(), endWindow.Month(), endWindow.Day()).Find(&occurrences)

		for _, occurrence := range occurrences {
			occurrenceDate := time.Date(now.Year(), time.Month(occurrence.Month), occurrence.Day, 0, 0, 0, 0, time.Local)
			if occurrenceDate.Before(now) || occurrenceDate.After(endWindow) {
				continue
			}

			if occurrence.Notify {
				notifyTelegram(occurrence)
				occurrence.Notified = true
				db.Save(&occurrence)
			}
		}

		// Check if New Year's Eve is within the next sleep cycle
		nextCheck := now.Add(sleepDuration)
		if now.Month() == 12 && now.Day() == 31 || (nextCheck.Month() == 1 && nextCheck.Day() == 1) {
			resetNotifications()
		}

		time.Sleep(sleepDuration)
	}
}
