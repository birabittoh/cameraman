package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"

	"github.com/gin-gonic/gin"
)

func padZero(i int) string {
	return fmt.Sprintf("%02d", i)
}

var (
	//go:embed templates/index.html
	templates     embed.FS
	indexTemplate *template.Template
	funcMap       = template.FuncMap{"padZero": padZero}
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
	db.Find(&occurrences)

	data := struct {
		Occurrences []Occurrence
	}{
		Occurrences: occurrences,
	}

	if indexTemplate.Execute(c.Writer, data) != nil {
		c.String(500, "Internal Server Error")
	}
}
