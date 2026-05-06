package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type DialogProps struct {
	BackdropID   string
	TitleID      string
	Title        string
	Header       g.Node
	Body         g.Node
	Actions      g.Node
	DialogClass  string
	BodyClass    string
	ActionsClass string
}

const DialogCSS = "/web/components/dialog.css"

func Dialog(props DialogProps) g.Node {
	titleID := props.TitleID
	if titleID == "" {
		titleID = "dialog-title"
	}

	return g.Group([]g.Node{
		InlineStyle(DialogCSS),
		h.Div(
			h.Class("dialog-backdrop is-hidden"),
			g.If(props.BackdropID != "", g.Attr("id", props.BackdropID)),
			h.Div(
				h.Class(dialogClasses(props.DialogClass)),
				g.Attr("role", "dialog"),
				g.Attr("aria-modal", "true"),
				g.Attr("aria-labelledby", titleID),
				g.If(props.Header != nil,
					props.Header,
				),
				g.If(props.Header == nil && props.Title != "",
					h.Div(
						h.Class("dialog-header"),
						h.H3(g.Attr("id", titleID), g.Text(props.Title)),
					),
				),
				g.If(props.Body != nil,
					h.Div(
						h.Class(bodyClasses(props.BodyClass)),
						props.Body,
					),
				),
				g.If(props.Actions != nil,
					h.Div(
						h.Class(actionClasses(props.ActionsClass)),
						props.Actions,
					),
				),
			),
		),
	})
}

func dialogClasses(extra string) string {
	if extra == "" {
		return "dialog"
	}
	return "dialog " + extra
}

func bodyClasses(extra string) string {
	if extra == "" {
		return "dialog-body"
	}
	return "dialog-body " + extra
}

func actionClasses(extra string) string {
	if extra == "" {
		return "dialog-actions"
	}
	return "dialog-actions " + extra
}
