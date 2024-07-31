package main

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
)

var md = goldmark.New()

func (p Post) Render() template.HTML {
	var out bytes.Buffer
	md.Convert([]byte(p.Content), &out)
	return template.HTML(out.String())
}

// @user -> username
// #post -> post

func (p Post) BuildReply() string {
	// naive solution: prefix everything?
	// User [@username](/u/username) wrote ...
	return p.Content
}
