package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/gin-gonic/gin"
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

func ShowIndexPage(c *gin.Context) {
	var occurrences []Occurrence
	db.Order("month, day, name").Find(&occurrences)

	data := struct {
		Occurrences []Occurrence
		CurrentYear uint
	}{
		Occurrences: occurrences,
		CurrentYear: uint(time.Now().Year()),
	}

	err := indexTemplate.Execute(c.Writer, data)
	if err != nil {
		log.Println(err.Error())
		c.String(500, "Internal Server Error")
	}
}
