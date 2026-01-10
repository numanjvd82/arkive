package web

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web/components"
)

type LayoutData struct {
	Title         string
	Description   string
	CanonicalURL  string
	Robots        string
	OGTitle       string
	OGDescription string
	OGImage       string
	OGType        string
	TwitterCard   string
	JSONLD        string
	CSS           []string
	JS            []string
	HideNav       bool
	User          *models.User
}

func Layout(data LayoutData, content ...g.Node) g.Node {
	pageTitle := data.Title
	if pageTitle == "" {
		pageTitle = "arkive.sh"
	}

	return h.Doctype(h.HTML(
		h.Lang("en"),
		h.Head(buildHeadNodes(LayoutData{
			Title:         pageTitle,
			Description:   data.Description,
			CanonicalURL:  data.CanonicalURL,
			Robots:        data.Robots,
			OGTitle:       data.OGTitle,
			OGDescription: data.OGDescription,
			OGImage:       data.OGImage,
			OGType:        data.OGType,
			TwitterCard:   data.TwitterCard,
			JSONLD:        data.JSONLD,
			CSS:           data.CSS,
			JS:            data.JS,
		})...),
		h.Body(
			g.If(!data.HideNav, components.Nav()),
			g.Group(content),
			components.ToastHost(),
		),
	))
}

func buildHeadNodes(data LayoutData) []g.Node {
	headNodes := []g.Node{
		h.Meta(h.Charset("utf-8")),
		h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
		h.Meta(h.Name("monetag"), h.Content("9d950a3ef0c449efafbbfa7840a36731")),
		h.TitleEl(g.Text(data.Title)),
		g.If(data.Description != "", h.Meta(h.Name("description"), h.Content(data.Description))),
		g.If(data.Robots != "", h.Meta(h.Name("robots"), h.Content(data.Robots))),
		g.If(data.CanonicalURL != "", h.Link(h.Rel("canonical"), h.Href(data.CanonicalURL))),
		g.If(data.OGTitle != "", h.Meta(g.Attr("property", "og:title"), h.Content(data.OGTitle))),
		g.If(data.OGDescription != "", h.Meta(g.Attr("property", "og:description"), h.Content(data.OGDescription))),
		g.If(data.OGType != "", h.Meta(g.Attr("property", "og:type"), h.Content(data.OGType))),
		g.If(data.OGImage != "", h.Meta(g.Attr("property", "og:image"), h.Content(data.OGImage))),
		g.If(data.CanonicalURL != "", h.Meta(g.Attr("property", "og:url"), h.Content(data.CanonicalURL))),
		g.If(data.TwitterCard != "", h.Meta(h.Name("twitter:card"), h.Content(data.TwitterCard))),
		g.If(data.OGTitle != "", h.Meta(h.Name("twitter:title"), h.Content(data.OGTitle))),
		g.If(data.OGDescription != "", h.Meta(h.Name("twitter:description"), h.Content(data.OGDescription))),
		g.If(data.OGImage != "", h.Meta(h.Name("twitter:image"), h.Content(data.OGImage))),
		h.Link(h.Rel("stylesheet"), h.Href("/static/reset.css")),
		h.Link(h.Rel("stylesheet"), h.Href("/static/globals.css")),
		components.InlineStyle(components.ButtonCSS),
		components.InlineStyle(components.ToastCSS),
		h.Script(
			h.Defer(),
			h.Src("https://static.cloudflareinsights.com/beacon.min.js"),
			g.Attr("data-cf-beacon", `{"token": "9ef8563310c5432ebe6f10581a7a6b7d"}`),
		),
		h.Link(h.Rel("icon"), h.Type("image/x-icon"), h.Href("/static/assets/images/favicon.ico")),
		h.Link(h.Rel("icon"), h.Type("image/png"), g.Attr("sizes", "32x32"), h.Href("/static/assets/images/favicon-32x32.png")),
		h.Link(h.Rel("icon"), h.Type("image/png"), g.Attr("sizes", "16x16"), h.Href("/static/assets/images/favicon-16x16.png")),
		h.Link(h.Rel("apple-touch-icon"), g.Attr("sizes", "180x180"), h.Href("/static/assets/images/apple-touch-icon.png")),
		h.Link(h.Rel("manifest"), h.Href("/static/assets/images/site.webmanifest")),
	}
	for _, css := range data.CSS {
		headNodes = append(headNodes, h.Link(h.Rel("stylesheet"), h.Href(css)))
	}
	headNodes = append(headNodes, h.Script(h.Src("/static/global.js"), h.Defer()))
	headNodes = append(headNodes, components.InlineScript(components.ToastJS))
	if data.JSONLD != "" {
		headNodes = append(headNodes, h.Script(h.Type("application/ld+json"), g.Raw(data.JSONLD)))
	}
	for _, src := range data.JS {
		headNodes = append(headNodes, h.Script(h.Src(src), h.Defer()))
	}
	return headNodes
}
