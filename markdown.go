package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

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
	var out bytes.Buffer
	// TODO link to post
	out.Write([]byte(fmt.Sprintf("[@%s](/u/%s) wrote:\n", p.Author.Username, p.Author.Username)))
	for _, line := range strings.Split(p.Content, "\n") {
		out.Write([]byte("> "))
		out.Write([]byte(line))
	}
	return out.String()
}
