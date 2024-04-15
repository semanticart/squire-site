package web

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"

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

func Render(data Page) []byte {
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

func Content(templateName string, page ...Page) template.HTML {
	var p Page
	if len(page) > 0 {
		p = page[0]
	} else {
		p = Page{}
	}

	return template.HTML(executeTemplateToBytes(templateName, p))
}

type Base64Download struct {
	Content string
	Name    string
	Format  string
	Error   bool
}

type UploadErrors struct {
	StoryErrors []squire.StoryError
	CustomError string
	Show        bool
}

type Page struct {
	Title      string
	Id         string
	Errors     UploadErrors
	Content    template.HTML
	CSSVersion string
	Download   Base64Download
}

func ReferrerIsNotSelf(r *http.Request) bool {
	referer := r.Header.Get("Referer")

	refererURL, err := url.Parse(referer)

	if err != nil {
		return false
	}

	return refererURL.Host != r.Host
}

func ValidateUpload(w http.ResponseWriter, r *http.Request) (string, error) {
	if ReferrerIsNotSelf(r) {
		return "", errors.New("Access denied")
	}

	if err := r.ParseMultipartForm(maxMemorySize); err != nil {
		return "", err
	}

	file, _, err := r.FormFile("fileUpload")
	if err != nil {
		return "", errors.New("Please choose a file to upload 🥺")
	}

	defer file.Close()

	content, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	return string(content), nil
}
