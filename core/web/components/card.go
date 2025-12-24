package components

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type CardProps struct {
	Title    string
	Subtitle string
	Class    string
	Body     []g.Node
}

func Card(props CardProps) g.Node {
	classValue := "card"
	if strings.TrimSpace(props.Class) != "" {
		classValue = classValue + " " + strings.TrimSpace(props.Class)
	}

	children := make([]g.Node, 0, 3)
	if props.Title != "" || props.Subtitle != "" {
		children = append(children, h.Div(
			h.Class("card-header"),
			g.If(props.Title != "", h.H3(g.Text(props.Title))),
			g.If(props.Subtitle != "", h.P(g.Text(props.Subtitle))),
		))
	}
	if len(props.Body) > 0 {
		children = append(children, h.Div(
			h.Class("card-body"),
			g.Group(props.Body),
		))
	}

	return h.Div(
		h.Class(classValue),
		g.Group(children),
	)
}
