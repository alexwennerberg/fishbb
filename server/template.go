package server

import (
	"embed"
	"html/template"
)

//go:embed views/* views/icons/*
var viewBundle embed.FS

var views *template.Template

func loadTemplates() *template.Template {
	views, err := template.New("main").Funcs(template.FuncMap{
		"timeago":     timeago,
		"addLinkTags": addLinkTags,
		"pageArr":     pageArray,
		"inc": func(i int) int {
			return i + 1

		},
		"dec": func(i int) int {
			return i - 1
		},
	}).ParseFS(viewBundle, "views/*.html", "views/icons/*.svg")
	if err != nil {
		panic(err)
	}
	return views
}
