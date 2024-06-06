package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const errJson = "{\"error\": %s}"

func returnError(w http.ResponseWriter, err error, status int) {
	http.Error(w, fmt.Sprintf(errJson, err.Error()), status)
}

func validateDate(month, day uint, year *uint) error {
	if month < 1 || month > 12 {
		return errors.New("invalid month: must be between 1 and 12")
	}

	if day < 1 || day > 31 {
		return errors.New("invalid day: must be between 1 and 31")
	}

	var testYear uint
	if year == nil {
		testYear = 2023
	} else {
		testYear = *year
	}

	// Construct a date and use time package to validate
	dateStr := fmt.Sprintf("%04d-%02d-%02d", testYear, month, day)
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New("invalid day for the given month")
	}

	return nil
}

func updateOccurrence(w http.ResponseWriter, input Occurrence) {
	var occurrence Occurrence
	if err := db.First(&occurrence, input.ID).Error; err != nil {
		returnError(w, err, http.StatusBadRequest)
		return
	}

	// Update existing record with new values
	if input.Day != 0 || input.Month != 0 {
		if err := validateDate(input.Month, input.Day, input.Year); err != nil {
			returnError(w, err, http.StatusBadRequest)
			return
		}
		occurrence.Day = input.Day
		occurrence.Month = input.Month
		occurrence.Year = input.Year
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
	json.NewEncoder(w).Encode(occurrence)
}

func addOccurrence(w http.ResponseWriter, r *http.Request) {
	var input Occurrence
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		returnError(w, err, http.StatusBadRequest)
		return
	}

	if input.ID != 0 {
		updateOccurrence(w, input)
		return
	}

	if err := validateDate(input.Month, input.Day, input.Year); err != nil {
		returnError(w, err, http.StatusBadRequest)
		return
	}

	db.Create(&input)
	json.NewEncoder(w).Encode(input)
}

func getOccurrences(w http.ResponseWriter, r *http.Request) {
	var occurrences []Occurrence
	db.Find(&occurrences)
	json.NewEncoder(w).Encode(occurrences)
}

func deleteOccurrence(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		returnError(w, err, http.StatusBadRequest)
		return
	}

	var occurrence Occurrence

	if err := db.First(&occurrence, id).Error; err != nil {
		returnError(w, err, http.StatusNotFound)
		return
	}

	db.Delete(&occurrence)
	w.WriteHeader(http.StatusNoContent)
}
