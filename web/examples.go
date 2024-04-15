package web

import "net/http"

var examplesTmpl = Render(Page{Title: "Squire - Examples", Id: "examples", Content: Content("examples.html.tmpl")})

func Examples(w http.ResponseWriter, r *http.Request) {
	w.Write(examplesTmpl)
}
