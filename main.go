package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/semanticart/squire/pkg"
)

//go:embed static/*
var staticFiles embed.FS

var t = template.Must(template.ParseFS(staticFiles, "static/*"))

const maxMemorySize = 10 << 20 // 10MB in bytes

func executeTemplateToBytes(templateName string, data Page) []byte {
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		panic("Could not execute template: " + err.Error())
	}
	return buf.Bytes()
}

type Page struct {
	Title      string
	Story      squire.Story
	ShowErrors bool
	Errors     []squire.StoryError
}

var empty = map[string]string{}

func main() {
	bookStyleSheet, err := staticFiles.ReadFile("static/style.css")
	if err != nil {
		panic("Could not read the file: " + err.Error())
	}

	index := executeTemplateToBytes("index.html.tmpl", Page{Title: "Squire", Story: squire.Story{}})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(index))
		case "POST":
			if err := r.ParseMultipartForm(maxMemorySize); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			file, _, err := r.FormFile("fileUpload")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			content, err := io.ReadAll(file)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			story, err := squire.ParseStory(string(content))

			var combinedStoryErrors *squire.CombinedStoryErrors
			errors.As(err, &combinedStoryErrors)

			error := executeTemplateToBytes("index.html.tmpl", Page{Title: "Squire", Story: story, Errors: combinedStoryErrors.Errors, ShowErrors: true})
			w.Write(error)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Write(bookStyleSheet)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":"+port, nil)
}
