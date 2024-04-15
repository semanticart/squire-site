package web

import "net/http"

var indexTmpl = Render(Page{Title: "Squire - Home", Id: "index", Content: Content("index.html.tmpl")})

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexTmpl)
}
