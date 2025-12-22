package web

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type LayoutData struct {
	Title string
	CSS   string
}

func Layout(data LayoutData, content ...g.Node) g.Node {
	headNodes := []g.Node{
		h.Meta(h.Charset("utf-8")),
		h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
		h.Title(data.Title),
		h.Link(h.Rel("stylesheet"), h.Href("/static/reset.css")),
		h.Link(h.Rel("stylesheet"), h.Href("/static/globals.css")),
	}
	if data.CSS != "" {
		headNodes = append(headNodes, h.Link(h.Rel("stylesheet"), h.Href(data.CSS)))
	}
	headNodes = append(headNodes,
		h.Script(h.Src("/static/global.js"),
			h.Defer(),
		))

	return h.Doctype(h.HTML(
		h.Lang("en"),
		h.Head(headNodes...),
		h.Body(
			g.Group(content),
		),
	))
}
