package components

import (
	"strconv"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type UploadInputProps struct {
	ID       string
	Name     string
	Label    string
	Helper   string
	Accept   string
	Required bool
}

func UploadInput(props UploadInputProps) g.Node {
	id := props.ID
	if id == "" {
		id = props.Name
	}
	input := h.Input(
		h.Type("file"),
		h.ID(id),
		h.Name(props.Name),
		h.Class("form-input upload-input"),
		g.If(props.Accept != "", g.Attr("accept", props.Accept)),
		g.If(props.Required, g.Attr("required", "required")),
	)

	helperID := ""
	if props.Helper != "" {
		helperID = id + "-helper"
		input = h.Input(
			h.Type("file"),
			h.ID(id),
			h.Name(props.Name),
			h.Class("form-input upload-input"),
			g.If(props.Accept != "", g.Attr("accept", props.Accept)),
			g.If(props.Required, g.Attr("required", "required")),
			g.Attr("aria-describedby", helperID),
		)
	}

	return h.Div(
		h.Class("form-field"),
		h.Label(
			h.Class("form-label"),
			h.For(id),
			g.Text(props.Label),
		),
		input,
		g.If(props.Helper != "",
			h.P(
				h.Class("form-subtitle"),
				h.ID(helperID),
				g.Text(props.Helper),
			),
		),
	)
}

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
		),
	)
}
