package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type inputType string

const (
	InputTypeText     inputType = "text"
	InputTypePassword inputType = "password"
	InputTypeEmail    inputType = "email"
	InputTypeNumber   inputType = "number"
)

type InputProps struct {
	Label       string
	Name        string
	Type        inputType
	Placeholder string
	Required    bool
}

func InputField(props InputProps) g.Node {
	inputType := props.Type
	if inputType == "" {
		inputType = "text"
	}
	id := props.Name

	inputClasses := "form-input"
	if inputType == "password" {
		inputClasses += " password-input"
	}

	input := h.Input(
		g.Attr("type", string(inputType)),
		g.Attr("name", props.Name),
		g.Attr("id", id),
		g.Attr("placeholder", props.Placeholder),
		g.If(
			props.Required,
			g.Attr("required", "required"),
		),
		h.Class(inputClasses),
	)

	if inputType == "password" {
		return h.Div(
			h.Class("form-field"),
			h.Label(
				h.Class("form-label"),
				g.Attr("for", id),
				g.Text(props.Label),
			),
			h.Div(
				h.Class("password-wrapper"),
				input,
				h.Button(
					h.Class("password-toggle"),
					g.Attr("type", "button"),
					g.Attr("aria-label", "Show password"),
					g.Attr("aria-pressed", "false"),
					g.Attr("data-target", id),
					g.Attr("data-visible", "false"),
					h.Img(
						h.Class("icon-eye"),
						g.Attr("src", "/static/icons/eye-open.svg"),
						g.Attr("alt", ""),
						g.Attr("aria-hidden", "true"),
					),
					h.Img(
						h.Class("icon-eye-off"),
						g.Attr("src", "/static/icons/eye-closed.svg"),
						g.Attr("alt", ""),
						g.Attr("aria-hidden", "true"),
					),
				),
			),
		)
	}

	return h.Div(
		h.Class("form-field"),
		h.Label(
			h.Class("form-label"),
			g.Attr("for", id),
			g.Text(props.Label),
		),
		input,
	)
}
