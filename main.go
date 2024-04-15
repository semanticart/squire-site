package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"

	web "github.com/semanticart/squire-site/web"
)

//go:embed web/static/*
var staticFiles embed.FS

func main() {
	fs := http.FS(staticFiles)
	fileServer := http.FileServer(fs)

	http.Handle("/web/static/site.css", http.StripPrefix("/", fileServer))

	http.HandleFunc("/", web.Index)
	http.HandleFunc("/validator", web.Validator)
	http.HandleFunc("/examples", web.Examples)
	http.HandleFunc("/converter", web.Converter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started on http://localhost:" + port + "/")
	http.ListenAndServe(":"+port, nil)
}
