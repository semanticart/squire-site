package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	squire "github.com/semanticart/squire/pkg"
)

//go:embed static/*
var staticFiles embed.FS

var t = template.Must(template.ParseFS(staticFiles, "static/*"))

const maxMemorySize = 10 << 20 // 10MB in bytes

func fingerprint(filename string) (string, error) {
	file, err := staticFiles.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:]), nil
}

func executeTemplateToBytes(templateName string, data Page) []byte {
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		panic("Could not execute template: " + err.Error())
	}
	return buf.Bytes()
}

var cssFingerprint string

func render(data Page) []byte {
	if cssFingerprint == "" {
		var err error
		cssFingerprint, err = fingerprint("static/site.css")
		if err != nil {
			log.Fatalf("Could not fingerprint site.css: %v", err)
		}
	}

	data.CSSVersion = cssFingerprint

	return executeTemplateToBytes("template.html.tmpl", data)
}

func content(templateName string, page ...Page) template.HTML {
	var p Page
	if len(page) > 0 {
		p = page[0]
	} else {
		p = Page{}
	}

	return template.HTML(executeTemplateToBytes(templateName, p))
}

type Page struct {
	Title       string
	Id          string
	ShowErrors  bool
	Errors      []squire.StoryError
	Content     template.HTML
	CustomError string
	CSSVersion  string
}

var pages = map[string]Page{
	"index":     {Title: "Squire - Home", Id: "index", Content: content("index.html.tmpl")},
	"converter": {Title: "Squire - Converter", Id: "converter", Content: content("converter.html.tmpl")},
	"examples":  {Title: "Squire - Examples", Id: "examples", Content: content("examples.html.tmpl")},
	"validator": {Title: "Squire - Validator", Id: "validator", Content: content("validator.html.tmpl")},
}

var renderedPages = map[string][]byte{
	"index":     render(pages["index"]),
	"converter": render(pages["converter"]),
	"examples":  render(pages["examples"]),
	"validator": render(pages["validator"]),
}

func renderUploadResponse(w http.ResponseWriter, errors []squire.StoryError, customError ...string) {
	var customErr string
	if len(customError) > 0 {
		customErr = customError[0]
	}

	combinedStoryErrors := squire.CombinedStoryErrors{Errors: errors}

	page := pages["validator"]
	page.CustomError = customErr
	page.Errors = combinedStoryErrors.Errors
	page.ShowErrors = true
	page.Content = content("validator.html.tmpl", page)

	result := render(page)

	w.Write(result)
}

func referrerIsNotSelf(r *http.Request) bool {
	referer := r.Header.Get("Referer")

	refererURL, err := url.Parse(referer)

	if err != nil {
		return false
	}

	return refererURL.Host != r.Host
}

func main() {
	fs := http.FS(staticFiles)
	fileServer := http.FileServer(fs)

	sum, err := fingerprint("static/site.css")
	if err != nil {
		log.Fatalf("Could not fingerprint site.css: %v", err)
	}

	fmt.Println("MD5 sum of site.css:", sum)

	http.Handle("/static/site.css", http.StripPrefix("/", fileServer))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(renderedPages["index"])
	})

	http.HandleFunc("/validator", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write(renderedPages["validator"])
		case "POST":
			if referrerIsNotSelf(r) {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

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

	http.HandleFunc("/examples", func(w http.ResponseWriter, r *http.Request) {
		w.Write(renderedPages["examples"])
	})

	http.HandleFunc("/converter", func(w http.ResponseWriter, r *http.Request) {
		w.Write(renderedPages["converter"])
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started on http://localhost:" + port + "/")
	http.ListenAndServe(":"+port, nil)
}
