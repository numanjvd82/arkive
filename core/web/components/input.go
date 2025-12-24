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
					h.Span(
						h.Class("icon-eye"),
						Icon(IconProps{
							Name:       "eye-open",
							Size:       "md",
							Decorative: true,
						}),
					),
					h.Span(
						h.Class("icon-eye-off"),
						Icon(IconProps{
							Name:       "eye-closed",
							Size:       "md",
							Decorative: true,
						}),
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
