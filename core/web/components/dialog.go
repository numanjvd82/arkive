package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type DialogProps struct {
	BackdropID string
	TitleID    string
	Title      string
	Body       g.Node
	Actions    g.Node
}

func Dialog(props DialogProps) g.Node {
	titleID := props.TitleID
	if titleID == "" {
		titleID = "dialog-title"
	}

	return h.Div(
		h.Class("dialog-backdrop is-hidden"),
		g.If(props.BackdropID != "", g.Attr("id", props.BackdropID)),
		h.Div(
			h.Class("dialog"),
			g.Attr("role", "dialog"),
			g.Attr("aria-modal", "true"),
			g.Attr("aria-labelledby", titleID),
			h.H3(g.Attr("id", titleID), g.Text(props.Title)),
			g.If(props.Body != nil, props.Body),
			g.If(props.Actions != nil, props.Actions),
		),
	)
}
