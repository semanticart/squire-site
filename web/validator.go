package web

import (
	"errors"
	"net/http"

	squire "github.com/semanticart/squire/pkg"
)

var validatorPage = Page{Title: "Squire - Validator", Id: "validator", Content: Content("validator.html.tmpl")}
var validatorTmpl = Render(validatorPage)

func renderValidatorUploadResponse(w http.ResponseWriter, errors []squire.StoryError, customError ...string) {
	var customErr string
	if len(customError) > 0 {
		customErr = customError[0]
	}

	combinedStoryErrors := squire.CombinedStoryErrors{Errors: errors}

	page := validatorPage
	page.Errors = UploadErrors{
		Show:        true,
		StoryErrors: combinedStoryErrors.Errors,
		CustomError: customErr,
	}
	page.Content = Content("validator.html.tmpl", page)

	result := Render(page)

	w.Write(result)
}

func Validator(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write(validatorTmpl)
	case "POST":
		content, err := ValidateUpload(w, r)

		if err != nil {
			renderValidatorUploadResponse(w, []squire.StoryError{}, err.Error())
			return
		}

		_, err = squire.ParseStory(string(content))

		if err != nil {
			var combinedStoryErrors *squire.CombinedStoryErrors
			errors.As(err, &combinedStoryErrors)

			renderValidatorUploadResponse(w, combinedStoryErrors.Errors)
		} else {
			renderValidatorUploadResponse(w, []squire.StoryError{})
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
