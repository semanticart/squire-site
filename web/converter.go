package web

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	squire "github.com/semanticart/squire/pkg"
)

var converterPage = Page{Title: "Squire - Converter", Id: "converter", Content: Content("converter.html.tmpl")}
var converterTmpl = Render(converterPage)

func renderConverterUploadResponse(w http.ResponseWriter, download Base64Download, errors []squire.StoryError, customError ...string) {
	var customErr string

	if len(customError) > 0 {
		customErr = customError[0]
	}

	combinedStoryErrors := squire.CombinedStoryErrors{Errors: errors}

	page := converterPage
	page.Errors = UploadErrors{
		Show:        true,
		StoryErrors: combinedStoryErrors.Errors,
		CustomError: customErr,
	}
	page.Download = download
	page.Content = Content("converter.html.tmpl", page)

	result := Render(page)

	w.Write(result)
}

func Converter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write(converterTmpl)
	case "POST":
		var buf bytes.Buffer
		content, err := ValidateUpload(w, r)

		if err != nil {
			renderConverterUploadResponse(w, Base64Download{}, []squire.StoryError{}, "Please choose a file to upload 🥺")
			return
		}

		story, parsingErrors := squire.ParseStory(string(content))

		format := "html"

		switch {
		case r.FormValue("format-epub") != "":
			format = "epub"
			buf, err = squire.ConvertToEpub(story)
		case r.FormValue("format-html") != "":
			buf, err = squire.ConvertToHtml(story, true)
		default:
			err = errors.New("No valid format specified")
			return
		}

		if err != nil || buf.Len() == 0 {
			renderConverterUploadResponse(w, Base64Download{}, []squire.StoryError{{
				Line: 0,
				Msg:  "Could not convert story",
			}})
			return
		}

		download := Base64Download{
			Content: base64.StdEncoding.EncodeToString(buf.Bytes()),
			Name:    "story." + format,
			Format:  strings.ToUpper(format),
		}

		if parsingErrors != nil {
			var combinedStoryErrors *squire.CombinedStoryErrors
			errors.As(parsingErrors, &combinedStoryErrors)

			download.Error = true

			renderConverterUploadResponse(w, download, combinedStoryErrors.Errors)
		} else {
			renderConverterUploadResponse(w, download, []squire.StoryError{})
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
