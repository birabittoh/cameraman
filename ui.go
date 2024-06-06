package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func calcYear(currentYear, year uint) uint {
	return currentYear - year
}

func padZero(i uint) string {
	return fmt.Sprintf("%02d", i)
}

var (
	//go:embed templates/index.html
	templates     embed.FS
	indexTemplate *template.Template
	funcMap       = template.FuncMap{
		"padZero":  padZero,
		"calcYear": calcYear,
	}
)

func ParseTemplates() {
	var err error
	indexTemplate, err = template.New("index.html").Funcs(funcMap).ParseFS(templates, "templates/index.html")
	if err != nil {
		log.Fatal("Could not parse index template")
		return
	}
}

func ShowIndexPage(w http.ResponseWriter, r *http.Request) {
	var occurrences []Occurrence
	db.Order("month, day, name").Find(&occurrences)

	data := struct {
		Occurrences []Occurrence
		CurrentYear uint
	}{
		Occurrences: occurrences,
		CurrentYear: uint(time.Now().Year()),
	}

	err := indexTemplate.Execute(w, data)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
