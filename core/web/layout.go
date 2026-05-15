package web

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web/components"
)

type LayoutData struct {
	Title              string
	Description        string
	CanonicalURL       string
	Robots             string
	OGTitle            string
	OGDescription      string
	OGImage            string
	OGType             string
	TwitterCard        string
	JSONLD             string
	CSS                []string
	JS                 []string
	ModuleJS           []string
	HideNav            bool
	User               *models.User
	ActiveNav          string
	SearchPlaceholder  string
	RequireVaultUnlock bool
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
			ModuleJS:      data.ModuleJS,
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
		h.Link(h.Rel("preconnect"), h.Href("https://fonts.googleapis.com")),
		h.Link(h.Rel("preconnect"), h.Href("https://fonts.gstatic.com"), g.Attr("crossorigin", "")),
		h.Link(
			h.Rel("stylesheet"),
			h.Href("https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap"),
		),
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
		h.Link(h.Rel("icon"), h.Type("image/x-icon"), h.Href("/static/assets/images/favicon.ico")),
		h.Link(h.Rel("icon"), h.Type("image/png"), g.Attr("sizes", "32x32"), h.Href("/static/assets/images/favicon-32x32.png")),
		h.Link(h.Rel("icon"), h.Type("image/png"), g.Attr("sizes", "48x48"), h.Href("/static/assets/images/favicon-48x48.png")),
		h.Link(h.Rel("icon"), h.Type("image/png"), g.Attr("sizes", "16x16"), h.Href("/static/assets/images/favicon-16x16.png")),
		h.Link(h.Rel("apple-touch-icon"), g.Attr("sizes", "180x180"), h.Href("/static/assets/images/apple-touch-icon.png")),
		h.Link(h.Rel("manifest"), h.Href("/static/assets/images/site.webmanifest")),
	}
	for _, css := range data.CSS {
		headNodes = append(headNodes, h.Link(h.Rel("stylesheet"), h.Href(css)))
	}
	headNodes = append(headNodes, h.Script(h.Type("module"), h.Src("/static/app.js")))
	if data.JSONLD != "" {
		headNodes = append(headNodes, h.Script(h.Type("application/ld+json"), g.Raw(data.JSONLD)))
	}
	for _, src := range data.JS {
		headNodes = append(headNodes, h.Script(h.Src(src), h.Defer()))
	}
	for _, src := range data.ModuleJS {
		headNodes = append(headNodes, h.Script(h.Type("module"), h.Src(src)))
	}
	return headNodes
}
