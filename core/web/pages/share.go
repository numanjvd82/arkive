package pages

import (
	"fmt"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type PublicSharePasswordProps struct {
	Token   string
	File    models.File
	Message string
}

func PublicSharePassword(props PublicSharePasswordProps) web.Page {
	file := props.File
	contentType := strings.TrimSpace(file.ContentType)
	if contentType == "" {
		contentType = "Unknown"
	}

	return web.Page{
		Title:   fmt.Sprintf("Arkive · %s", file.Filename),
		CSS:     []string{"/web/pages/share.css"},
		JS:      []string{"/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		HideNav: true,
		Body: g.Group([]g.Node{
			components.InlineStyle(components.InputCSS),
			h.Main(
				h.Class("share-page"),
				h.Div(
					h.Class("share-card"),
					h.Div(
						h.Class("share-header"),
						h.P(h.Class("share-eyebrow"), g.Text("Private link")),
						h.H1(g.Text("Enter password to view")),
						h.P(
							h.Class("share-subtitle"),
							g.Text("This file is protected. Enter the password to continue."),
						),
					),
					h.Div(
						h.Class("share-file-meta"),
						metaRow("Filename", file.Filename),
						metaRow("Type", contentType),
						metaRow("Size", format.Bytes(file.SizeBytes)),
					),
					h.Form(
						h.Class("share-form"),
						h.Method("POST"),
						h.Action("/s/"+props.Token),
						h.Div(
							h.Class("form-field"),
							h.Label(
								h.Class("form-label"),
								g.Attr("for", "share-password"),
								g.Text("Password"),
							),
							h.Div(
								h.Class("password-wrapper"),
								h.Input(
									h.Class("form-input password-input"),
									g.Attr("id", "share-password"),
									g.Attr("name", "password"),
									g.Attr("type", "password"),
									g.Attr("placeholder", "Enter password"),
									g.Attr("required", "required"),
								),
								h.Button(
									h.Class("password-toggle"),
									g.Attr("type", "button"),
									g.Attr("aria-label", "Show password"),
									g.Attr("aria-pressed", "false"),
									g.Attr("data-target", "share-password"),
									g.Attr("data-visible", "false"),
									h.Span(
										h.Class("icon-eye"),
										components.Icon(components.IconProps{
											Name:       "eye-open",
											Size:       "md",
											Decorative: true,
										}),
									),
									h.Span(
										h.Class("icon-eye-off"),
										components.Icon(components.IconProps{
											Name:       "eye-closed",
											Size:       "md",
											Decorative: true,
										}),
									),
								),
							),
						),
						g.If(props.Message != "", h.P(
							h.Class("form-error"),
							g.Text(props.Message),
						)),
						h.Button(
							h.Class("button primary share-submit"),
							h.Type("submit"),
							g.Text("Unlock file"),
						),
					),
				),
			),
		}),
	}
}
