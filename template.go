package main

import (
	"embed"
	"html/template"
	"time"
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
		"isodate": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format(time.RFC3339)
		},
	}).ParseFS(viewBundle, "views/*.html", "views/icons/*.svg")
	if err != nil {
		panic(err)
	}
	return views
}
