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
	Title       string
	ShowErrors  bool
	Errors      []squire.StoryError
	CustomError string
}

var empty = map[string]string{}

func renderUploadResponse(w http.ResponseWriter, errors []squire.StoryError, customError ...string) {
	var customErr string
	if len(customError) > 0 {
		customErr = customError[0]
	}

	combinedStoryErrors := squire.CombinedStoryErrors{Errors: errors}

	error := executeTemplateToBytes("index.html.tmpl", Page{Title: "Squire", Errors: combinedStoryErrors.Errors, ShowErrors: true, CustomError: customErr})
	w.Write(error)
}

func main() {
	bookStyleSheet, err := staticFiles.ReadFile("static/style.css")
	if err != nil {
		panic("Could not read the file: " + err.Error())
	}

	index := executeTemplateToBytes("index.html.tmpl", Page{Title: "Squire"})

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
				renderUploadResponse(w, []squire.StoryError{}, "Please choose a file to upload 🥺")
				return
			}
			defer file.Close()

			content, err := io.ReadAll(file)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = squire.ParseStory(string(content))

			if err != nil {
				var combinedStoryErrors *squire.CombinedStoryErrors
				errors.As(err, &combinedStoryErrors)

				renderUploadResponse(w, combinedStoryErrors.Errors)
			} else {
				renderUploadResponse(w, []squire.StoryError{})
			}
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
