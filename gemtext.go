package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"git.sr.ht/~aw/gmi2html"
)

func (p Post) Render() template.HTML {
	r := gmi2html.NewReader(strings.NewReader(p.Content))
	r.NestedBlocks = true
	return template.HTML(r.HTMLString())
}

func (p Post) BuildReply() string {
	var out bytes.Buffer
	out.Write([]byte(fmt.Sprintf("@%s wrote:\n", p.Author.Username)))
	for _, line := range strings.Split(p.Content, "\n") {
		out.Write([]byte(">"))
		out.Write([]byte(line))
	}
	out.Write([]byte("\n"))
	return out.String()
}
