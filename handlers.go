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

func addOccurrence(c *gin.Context) {
	var input Occurrence
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate date
	if err := validateDate(input.Month, input.Day); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var occurrence Occurrence
	if input.ID != 0 {
		if err := db.First(&occurrence, input.ID).Error; err == nil {
			// Update existing record with new values
			if input.Month != 0 {
				occurrence.Month = input.Month
			}
			if input.Day != 0 {
				occurrence.Day = input.Day
			}
			if input.Name != "" {
				occurrence.Name = input.Name
			}
			if input.Description != "" {
				occurrence.Description = input.Description
			}
			occurrence.Notify = input.Notify
			db.Save(&occurrence)
			c.JSON(http.StatusOK, occurrence)
			return
		}
	}

	// Create a new record if no existing record is found
	occurrence = input
	db.Create(&occurrence)
	c.JSON(http.StatusOK, occurrence)
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
