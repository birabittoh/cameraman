package main

import (
	"fmt"
	"html"
	"log"
	"sort"
	"strings"
	"time"
)

// Telegram rejects messages longer than 4096 characters.
const telegramMaxMessageLength = 4096

var monthNames = [...]string{
	"Gennaio", "Febbraio", "Marzo", "Aprile", "Maggio", "Giugno",
	"Luglio", "Agosto", "Settembre", "Ottobre", "Novembre", "Dicembre",
}

// MonthlySummary records which months already had their summary sent.
type MonthlySummary struct {
	Year   int `gorm:"primaryKey;autoIncrement:false"`
	Month  int `gorm:"primaryKey;autoIncrement:false"`
	SentAt time.Time
}

// formatMonthlySummary builds HTML-formatted Telegram messages listing the
// given occurrences, split so that each one fits Telegram's length limit.
func formatMonthlySummary(year int, month time.Month, occurrences []Occurrence) []string {
	sort.SliceStable(occurrences, func(i, j int) bool {
		return occurrences[i].Day < occurrences[j].Day
	})

	header := fmt.Sprintf("📅 <b>Riepilogo di %s %d</b>\n", monthNames[month-1], year)

	var messages []string
	var b strings.Builder
	b.WriteString(header)

	for _, o := range occurrences {
		var entry strings.Builder
		entry.WriteString(fmt.Sprintf("\n<b>%02d/%02d</b> - <i>%s</i>", o.Day, o.Month, html.EscapeString(o.Name)))
		if o.Year != nil {
			entry.WriteString(fmt.Sprintf(" 🎂 %d anni", year-int(*o.Year)))
		}
		if d := strings.TrimSpace(o.Description); d != "" {
			entry.WriteString("\n" + html.EscapeString(d))
		}
		entry.WriteString("\n")

		if b.Len()+entry.Len() > telegramMaxMessageLength && b.Len() > len(header) {
			messages = append(messages, b.String())
			b.Reset()
			b.WriteString(header)
		}
		b.WriteString(entry.String())
	}

	return append(messages, b.String())
}

// checkMonthlySummary sends the summary of the current month's occurrences,
// once per month.
func checkMonthlySummary() {
	now := time.Now()
	year, month := now.Year(), now.Month()

	var count int64
	if err := db.Model(&MonthlySummary{}).Where("year = ? AND month = ?", year, int(month)).Count(&count).Error; err != nil {
		log.Printf("Failed to check monthly summary: %v", err)
		return
	}
	if count > 0 {
		return
	}

	var occurrences []Occurrence
	if err := db.Where("month = ?", int(month)).Find(&occurrences).Error; err != nil {
		log.Printf("Failed to load occurrences for monthly summary: %v", err)
		return
	}

	if len(occurrences) > 0 {
		log.Println("Sending monthly summary for", month, year)
		for _, message := range formatMonthlySummary(year, month, occurrences) {
			if _, err := sendTelegramMessage(message, "HTML"); err != nil {
				log.Printf("Failed to send monthly summary: %v", err)
				return
			}
		}
	}

	if err := db.Create(&MonthlySummary{Year: year, Month: int(month), SentAt: now}).Error; err != nil {
		log.Printf("Failed to save monthly summary: %v", err)
	}
}
