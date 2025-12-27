package components

import (
	"strconv"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type ProgressBarProps struct {
	ID    string
	Value int
	Label string
}

func ProgressBar(props ProgressBarProps) g.Node {
	value := props.Value
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}

	return h.Div(
		h.Class("progress"),
		g.If(props.ID != "", g.Attr("id", props.ID)),
		g.If(props.Label != "",
			h.Span(h.Class("progress-label"), g.Text(props.Label)),
		),
		h.Div(
			h.Class("progress-track"),
			h.Div(
				h.Class("progress-bar"),
				g.Attr("style", "width: "+strconv.Itoa(value)+"%"),
				g.Attr("role", "progressbar"),
				g.Attr("aria-valuemin", "0"),
				g.Attr("aria-valuemax", "100"),
				g.Attr("aria-valuenow", strconv.Itoa(value)),
			),
			h.Span(h.Class("progress-percent"), g.Text(strconv.Itoa(value)+"%")),
		),
	)
}
