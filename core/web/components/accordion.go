package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type AccordionProps struct {
	ID      string
	Class   string
	Summary g.Node
	Body    g.Node
	Open    bool
}

const AccordionCSS = "/web/components/accordion.css"

func Accordion(props AccordionProps) g.Node {
	className := "accordion"
	if props.Class != "" {
		className += " " + props.Class
	}

	return g.Group([]g.Node{
		InlineStyle(AccordionCSS),
		h.Details(
			h.Class(className),
			g.If(props.ID != "", g.Attr("id", props.ID)),
			g.If(props.Open, g.Attr("open", "open")),
			h.Summary(
				h.Class("accordion-summary"),
				props.Summary,
			),
			h.Div(
				h.Class("accordion-body"),
				props.Body,
			),
		),
	})
}
