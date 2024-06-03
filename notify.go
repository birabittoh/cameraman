package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	NotificationWindow int
	SleepDuration      time.Duration

	notificationClient *resty.Client
	telegramToken      string
	chatID             string
	threadID           string
)

func notifyTelegram(occurrence Occurrence) error {
	log.Println("Sending notification for occurrence", occurrence.ID)
	message := fmt.Sprintf("*Giorno %02d/%02d*.\n\n_%s_\n%s",
		occurrence.Day, occurrence.Month, occurrence.Name, occurrence.Description)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramToken)

	// Create the payload
	payload := map[string]interface{}{
		"chat_id":           chatID,
		"text":              message,
		"parse_mode":        "markdown",
		"message_thread_id": threadID,
	}

	// Send the POST request
	resp, err := notificationClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(url)

	if err != nil {
		log.Printf("Failed to send notification: %v", err)
		return err
	}
	log.Printf("Notification sent: %s, Response: %s", message, resp)
	return nil
}

func resetNotifications() {
	if err := db.Model(&Occurrence{}).Where("notified = ?", true).Update("notified", false).Error; err != nil {
		log.Printf("Failed to reset notifications: %v", err)
	} else {
		log.Println("Notifications have been reset for the new year.")
	}
}

func initNotifications() error {
	notificationClient = resty.New()

	telegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID = os.Getenv("TELEGRAM_CHAT_ID")
	threadID = os.Getenv("TELEGRAM_THREAD_ID")

	if telegramToken == "" || chatID == "" {
		log.Println("Warning: you should set your Telegram Bot token and chat id in .env, otherwise you won't get notifications!")
		return errors.New("empty telegramToken or chatId")
	}
	return nil
}

func CheckOccurrences() {
	err := initNotifications()
	if err != nil {
		log.Println(err.Error())
		return
	}

	for {
		log.Println("Checking for new occurrences.")
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		endWindow := today.AddDate(0, 0, NotificationWindow)

		var occurrences []Occurrence
		db.Where("notified = ? AND ((month = ? AND day >= ?) OR (month = ? AND day <= ?))",
			false, today.Month(), today.Day(), endWindow.Month(), endWindow.Day()).Find(&occurrences)

		for _, occurrence := range occurrences {
			occurrenceDate := time.Date(today.Year(), time.Month(occurrence.Month), int(occurrence.Day), 0, 0, 0, 0, time.Local)
			if occurrenceDate.Before(today) || occurrenceDate.After(endWindow) || !occurrence.Notify || occurrence.Notified {
				continue
			}

			err := notifyTelegram(occurrence)
			if err != nil {
				log.Println(err.Error())
				return
			}
			occurrence.Notified = true
			db.Save(&occurrence)
		}

		// Check if New Year's Eve is within the next sleep cycle
		nextCheck := now.Add(SleepDuration)
		if (now.Month() == 12 && now.Day() == 31) && (nextCheck.Month() == 1 && nextCheck.Day() == 1) {
			resetNotifications()
		}

		time.Sleep(SleepDuration)
	}
}
