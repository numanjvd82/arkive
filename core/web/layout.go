package web

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web/components"
)

type LayoutData struct {
	Title string
	CSS   []string
	JS    []string
}

func Layout(data LayoutData, content ...g.Node) g.Node {
	headNodes := []g.Node{
		h.Meta(h.Charset("utf-8")),
		h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
		h.Title(data.Title),
		h.Link(h.Rel("stylesheet"), h.Href("/static/reset.css")),
		h.Link(h.Rel("stylesheet"), h.Href("/static/globals.css")),
		h.Link(h.Rel("stylesheet"), h.Href("/static/components.css")),
	}
	for _, css := range data.CSS {
		headNodes = append(headNodes, h.Link(h.Rel("stylesheet"), h.Href(css)))
	}
	headNodes = append(headNodes, h.Script(h.Src("/static/global.js"), h.Defer()))
	for _, src := range data.JS {
		headNodes = append(headNodes, h.Script(h.Src(src), h.Defer()))
	}

	return h.Doctype(h.HTML(
		h.Lang("en"),
		h.Head(headNodes...),
		h.Body(
			components.Nav(),
			g.Group(content),
		),
	))
}
