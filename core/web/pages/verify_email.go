package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type VerifyEmailPageProps struct {
	Ctx     PageContext
	Success bool
	Message string
}

func VerifyEmailPage(props VerifyEmailPageProps) web.Page {
	return web.Page{
		Title:   "Arkive · Verify Email",
		Robots:  RobotsNoIndex,
		CSS:     []string{"/web/pages/verify_email.css"},
		Body:    verifyEmailBody(props),
		HideNav: true,
	}
}

func verifyEmailBody(props VerifyEmailPageProps) g.Node {
	title := "Email verified"
	if !props.Success {
		title = "Verify email"
	}

	message := props.Message
	if message == "" {
		if props.Success {
			message = "Your email is confirmed. You can log in now."
		} else {
			message = "This verification link is invalid or expired."
		}
	}

	return h.Div(
		h.Class("auth-page"),
		h.Div(
			h.Class("auth-container"),
			components.Card(components.CardProps{
				Title:    title,
				Subtitle: message,
				Class:    "auth-card",
				Body: []g.Node{
					h.Div(
						h.Class("verify-actions"),
						components.Button(components.ButtonProps{
							Text:    "Go to login",
							Href:    "/login",
							Variant: "primary",
							Class:   "auth-submit",
						}),
						components.Button(components.ButtonProps{
							Text:    "Back to home",
							Href:    "/",
							Variant: "secondary",
							Class:   "auth-submit",
						}),
					),
				},
			}),
		),
	)
}
