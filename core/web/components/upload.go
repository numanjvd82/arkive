package components

import (
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

const UploadInputCSS = "/web/components/upload_input.css"

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

	return g.Group([]g.Node{
		InlineStyle(UploadInputCSS),
		h.Div(
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
		),
	})
}
