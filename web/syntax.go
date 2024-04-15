package web

import "net/http"

var syntaxTmpl = Render(Page{Title: "Squire - Syntax", Id: "syntax", Content: Content("syntax.html.tmpl")})

func Syntax(w http.ResponseWriter, r *http.Request) {
	w.Write(syntaxTmpl)
}
