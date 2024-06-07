package main

import (
	"html/template"
	"net/http"
)

var config config

func indexHandler(w http.ResponseWriter, r *http.Request) {
}

func generateTemplates() {
	templateFuncs := template.FuncMap{}
}

func serve() {
	config = NewConfig()
	db = opendb()
	prepareStatements(db)
	// Setup Templates
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.ListenAndServe(config.port)
}

func main() {
	serve()
}
