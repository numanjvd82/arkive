package components

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type Breadcrumb struct {
	Title    string
	Href     string
	IconName string
}

type BreadcrumbsProps struct {
	Items []Breadcrumb
	Class string
}

const BreadcrumbsCSS = "/web/components/breadcrumbs.css"

func Breadcrumbs(props BreadcrumbsProps) g.Node {
	if len(props.Items) == 0 {
		return nil
	}

	classes := "breadcrumbs"
	if strings.TrimSpace(props.Class) != "" {
		classes += " " + strings.TrimSpace(props.Class)
	}

	lastIndex := len(props.Items) - 1
	crumbNodes := make([]g.Node, 0, len(props.Items))
	for idx, item := range props.Items {
		title := strings.TrimSpace(item.Title)
		crumbContent := g.Group([]g.Node{
			g.If(strings.TrimSpace(item.IconName) != "", Icon(IconProps{
				Name:       item.IconName,
				Size:       "18",
				Decorative: true,
			})),
			h.Span(g.Text(title)),
		})

		link := strings.TrimSpace(item.Href)
		isLast := idx == lastIndex || link == ""
		crumbNodes = append(crumbNodes, h.Li(
			g.If(idx > 0, h.Span(h.Class("divider"), g.Text("/"))),
			g.If(isLast, h.Span(h.Class("crumb"), crumbContent)),
			g.If(!isLast, h.A(
				h.Href(link),
				h.Class("crumb"),
				crumbContent,
			)),
		))
	}
	items := g.Group(crumbNodes)

	return g.Group([]g.Node{
		InlineStyle(BreadcrumbsCSS),
		h.Nav(
			h.Class(classes),
			h.Aria("label", "breadcrumbs"),
			h.Ol(
				h.Class("breadcrumbs-list"),
				items,
			),
		),
	})
}
