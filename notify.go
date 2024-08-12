package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const msgFormat = "*Giorno %02d/%02d*\n\n_%s_\n%s"
const msgFormatYear = "*Giorno %02d/%02d/%04d*\n🎂 %d anni\n\n_%s_\n%s"
const baseUrl = "https://api.telegram.org/bot"

type Response struct {
	Ok     bool   `json:"ok"`
	Result Result `json:"result"`
}

type Result struct {
	MessageID int `json:"message_id"`
}

var (
	NotificationWindow     int
	SoftNotificationWindow int
	SleepDuration          time.Duration
	telegramToken          string
	chatID                 string
	threadID               string
)

func sendPostRequest(url string, payload map[string]interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	return resp, nil
}

func notifyTelegram(occurrence Occurrence, soft bool) error {
	log.Println("Sending notification for occurrence", occurrence.ID)
	var message string
	if occurrence.Year != nil {
		years := time.Now().Year() - int(*occurrence.Year)
		message = fmt.Sprintf(msgFormatYear, occurrence.Day, occurrence.Month, *occurrence.Year, years, occurrence.Name, occurrence.Description)
	} else {
		message = fmt.Sprintf(msgFormat, occurrence.Day, occurrence.Month, occurrence.Name, occurrence.Description)
	}

	url := fmt.Sprintf("%s%s/sendMessage", baseUrl, telegramToken)

	// Create the payload
	payload := map[string]interface{}{
		"chat_id":              chatID,
		"text":                 message,
		"parse_mode":           "markdown",
		"message_thread_id":    threadID,
		"disable_notification": true,
	}

	// Send the POST request
	resp, err := sendPostRequest(url, payload)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return err
	}

	log.Printf("Notification sent: %s, Response: %s", message, string(bodyBytes))

	// Decode the JSON response
	var r Response
	if err := json.Unmarshal(bodyBytes, &r); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return err
	}

	if !r.Ok {
		log.Printf("Telegram API returned an error: %v", r)
		return fmt.Errorf("telegram API error: %v", r)
	}

	if soft {
		return nil
	}

	msgId := r.Result.MessageID

	// Prepare the request to pin the message
	url = fmt.Sprintf("%s%s/pinChatMessage", baseUrl, telegramToken)
	payload = map[string]interface{}{
		"chat_id":              chatID,
		"message_id":           msgId,
		"disable_notification": false,
	}

	// Send the POST request to pin the message
	resp, err = sendPostRequest(url, payload)
	if err != nil {
		log.Printf("Failed to pin message: %v", err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return err
	}

	log.Printf("Message pinned: %s, Response: %s", message, string(bodyBytes))
	return nil
}

func resetNotifications(notified_column string) (err error) {
	if err = db.Model(&Occurrence{}).Where(notified_column+" = ?", true).Update(notified_column, false).Error; err != nil {
		log.Printf("Failed to reset notifications: %v", err)
	} else {
		log.Println("Notifications have been reset for the new year.")
	}
	return
}

func initNotifications() error {
	telegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID = os.Getenv("TELEGRAM_CHAT_ID")
	threadID = os.Getenv("TELEGRAM_THREAD_ID")

	if telegramToken == "" || chatID == "" {
		log.Println("Warning: you should set your Telegram Bot token and chat id in .env, otherwise you won't get notifications!")
		return errors.New("empty telegramToken or chatId")
	}
	return nil
}

func check(notificationWindowDays int, soft bool) (err error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	endWindow := today.AddDate(0, 0, notificationWindowDays)

	notified_column := "notified"
	if soft {
		notified_column += "_soft"
	}

	var occurrences []Occurrence
	db.Where(notified_column+" = ? AND ((month = ? AND day >= ?) OR (month = ? AND day <= ?))",
		false, today.Month(), today.Day(), endWindow.Month(), endWindow.Day()).Find(&occurrences)

	for _, occurrence := range occurrences {
		occurrenceDate := time.Date(today.Year(), time.Month(occurrence.Month), int(occurrence.Day), 0, 0, 0, 0, time.Local)
		if occurrenceDate.Before(today) || occurrenceDate.After(endWindow) || !occurrence.Notify || occurrence.Notified {
			continue
		}

		err = notifyTelegram(occurrence, soft)
		if err != nil {
			return err
		}

		err = db.Model(&Occurrence{}).Where("id = ?", occurrence.ID).Update(notified_column, true).Error
		if err != nil {
			return err
		}
	}

	// Check if New Year's Eve is within the next sleep cycle
	nextCheck := now.Add(SleepDuration)
	if (now.Month() == 12 && now.Day() == 31) && (nextCheck.Month() == 1 && nextCheck.Day() == 1) {
		resetNotifications(notified_column)
	}

	return
}

func CheckOccurrences() {
	err := initNotifications()
	if err != nil {
		log.Println(err.Error())
		return
	}

	for {
		log.Println("Checking for new occurrences.")

		check(NotificationWindow, false)
		check(SoftNotificationWindow, true)

		time.Sleep(SleepDuration)
	}
}
