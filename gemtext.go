package main

import (
	"html/template"
	"strings"

	"git.sr.ht/~aw/gmi2html"
)

func (p Post) Render() template.HTML {
	r := strings.NewReader(p.Content)
	g := gmi2html.NewReader(r)
	g.NestedBlocks = true
	return template.HTML(g.HTMLString())
}
