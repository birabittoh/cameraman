package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func validateDate(month, day int) error {
	if month < 1 || month > 12 {
		return errors.New("invalid month: must be between 1 and 12")
	}

	if day < 1 || day > 31 {
		return errors.New("invalid day: must be between 1 and 31")
	}

	// Construct a date and use time package to validate
	dateStr := fmt.Sprintf("2023-%02d-%02d", month, day)
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New("invalid day for the given month")
	}

	return nil
}

func updateOccurrence(c *gin.Context, input Occurrence) {
	var occurrence Occurrence
	if err := db.First(&occurrence, input.ID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update existing record with new values
	if input.Day != 0 || input.Month != 0 {
		if err := validateDate(input.Month, input.Day); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		occurrence.Month = input.Month
		occurrence.Day = input.Day
	}
	if input.Name != "" {
		occurrence.Name = input.Name
	}
	if input.Description != "" {
		occurrence.Description = input.Description
	}
	occurrence.Notify = input.Notify
	occurrence.Notified = input.Notified
	db.Save(&occurrence)
	c.JSON(http.StatusOK, occurrence)
}

func addOccurrence(c *gin.Context) {
	var input Occurrence
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ID != 0 {
		updateOccurrence(c, input)
		return
	}

	if err := validateDate(input.Month, input.Day); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Create(&input)
	c.JSON(http.StatusOK, input)
}

func getOccurrences(c *gin.Context) {
	var occurrences []Occurrence
	db.Find(&occurrences)
	c.JSON(http.StatusOK, occurrences)
}

func deleteOccurrence(c *gin.Context) {
	id := c.Param("id")
	var occurrence Occurrence

	if err := db.First(&occurrence, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Occurrence not found"})
		return
	}

	db.Delete(&occurrence)
	c.JSON(http.StatusOK, gin.H{"message": "Occurrence deleted"})
}
