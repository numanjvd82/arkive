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
	Description string
	Name        string
	Type        inputType
	Placeholder string
	Value       string
	Required    bool
	HelperText  string
	HasError    bool
}

const InputCSS = "/web/components/input.css"

func InputField(props InputProps) g.Node {
	inputType := props.Type
	if inputType == "" {
		inputType = "text"
	}
	id := props.Name

	hasError := props.HasError
	if props.HelperText != "" && !props.HasError {
		hasError = true
	}

	inputClasses := "form-input"
	if inputType == "password" {
		inputClasses += " password-input"
	}
	if hasError {
		inputClasses += " is-invalid"
	}

	helperID := ""
	if props.HelperText != "" {
		helperID = id + "-helper"
	}

	input := h.Input(
		g.Attr("type", string(inputType)),
		g.Attr("name", props.Name),
		g.Attr("id", id),
		g.Attr("placeholder", props.Placeholder),
		g.If(
			props.Value != "",
			g.Attr("value", props.Value),
		),
		g.If(
			props.Required,
			g.Attr("required", "required"),
		),
		g.If(
			hasError,
			g.Attr("aria-invalid", "true"),
		),
		g.If(
			helperID != "",
			g.Attr("aria-describedby", helperID),
		),
		h.Class(inputClasses),
	)

	var node g.Node
	if inputType == "password" {
		node = h.Div(
			h.Class("form-field"),
			h.Label(
				h.Class("form-label"),
				g.Attr("for", id),
				g.Text(props.Label),
			),
			g.If(
				props.Description != "",
				h.P(
					h.Class("form-subtitle"),
					g.Text(props.Description),
				),
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
			g.If(
				props.HelperText != "",
				h.P(
					h.Class("form-helper"),
					g.Attr("id", helperID),
					g.Text(props.HelperText),
				),
			),
		)
	} else {
		node = h.Div(
			h.Class("form-field"),
			h.Label(
				h.Class("form-label"),
				g.Attr("for", id),
				g.Text(props.Label),
			),
			g.If(
				props.Description != "",
				h.P(
					h.Class("form-subtitle"),
					g.Text(props.Description),
				),
			),
			input,
			g.If(
				props.HelperText != "",
				h.P(
					h.Class("form-helper"),
					g.Attr("id", helperID),
					g.Text(props.HelperText),
				),
			),
		)
	}

	return g.Group([]g.Node{
		InlineStyle(InputCSS),
		node,
	})
}
