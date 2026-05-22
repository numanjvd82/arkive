package pages

import (
	"fmt"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type NotFoundPageProps struct {
	Ctx  PageContext
	Path string
}

func NotFoundPage(props NotFoundPageProps) web.Page {
	title := "Arkive · Page not found"
	if props.Path != "" {
		title = fmt.Sprintf("Arkive · %s not found", props.Path)
	}

	return web.Page{
		Title:   title,
		Robots:  RobotsNoIndex,
		CSS:     []string{"/web/pages/not_found.css"},
		HideNav: true,
		User:    props.Ctx.User,
		Body: h.Main(
			h.Class("not-found"),
			h.Div(h.Class("not-found-grid-overlay")),
			h.Div(
				h.Class("not-found-shell"),
				h.Div(
					h.Class("not-found-visual"),
					h.Div(
						h.Class("not-found-visual-card"),
						lucide.Unlink(
							h.Class("not-found-visual-icon"),
							g.Attr("aria-hidden", "true"),
						),
						h.Div(h.Class("not-found-corner not-found-corner-top-left")),
						h.Div(h.Class("not-found-corner not-found-corner-top-right")),
						h.Div(h.Class("not-found-corner not-found-corner-bottom-left")),
						h.Div(h.Class("not-found-corner not-found-corner-bottom-right")),
					),
					h.Div(
						h.Class("not-found-code-row"),
						h.Span(h.Class("not-found-code-dot")),
						h.Span(h.Class("not-found-code-text"), g.Text("ERROR_CODE: PATH_UNRESOLVED")),
					),
				),
				h.Section(
					h.Class("not-found-copy"),
					h.H1(
						h.Class("not-found-title"),
						g.Text("404: Node Path Not Found"),
					),
					h.P(
						h.Class("not-found-lead"),
						g.Text("The requested resource does not exist on this node or has been moved. The address may be invalid or no longer available."),
					),
					h.Div(
						h.Class("not-found-actions"),
						components.Button(components.ButtonProps{
							Text:    "Return to Dashboard",
							Href:    "/dashboard",
							Variant: "primary",
							Class:   "not-found-action-primary",
						}),
						components.Button(components.ButtonProps{
							Text:    "Open Files",
							Href:    "/files",
							Variant: "secondary",
							Class:   "not-found-action-secondary",
							Icon:    "search",
						}),
					),
					h.Div(
						h.Class("not-found-footer"),
						h.P(
							h.Class("not-found-footer-copy"),
							g.Text("Powered by "),
							h.Span(h.Class("not-found-footer-strong"), g.Text("Arkive Core")),
						),
						lucide.BadgeCheck(
							h.Class("not-found-footer-icon"),
							g.Attr("aria-hidden", "true"),
						),
					),
				),
			),
		),
	}
}
