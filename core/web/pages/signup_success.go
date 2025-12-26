package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

func SignupSuccessPage() web.Page {
	return web.Page{
		Title: "arkive.sh · Welcome",
		CSS:   []string{"/web/pages/signup_success.css"},
		JS: []string{
			"https://cdn.jsdelivr.net/npm/canvas-confetti@1.6.0/dist/confetti.browser.min.js",
			"/static/signup_success.js",
		},
		Body: signupSuccessBody(),
	}
}

func signupSuccessBody() g.Node {
	return h.Div(
		h.Class("success-page"),
		h.Div(
			h.Class("success-container"),
			components.Card(components.CardProps{
				Title:    "You are in",
				Subtitle: "Welcome to arkive.sh",
				Class:    "success-card",
				Body: []g.Node{
					h.P(
						h.Class("success-text"),
						g.Text("When the app launches, log in with this account and get 5GB of free space, early access to new features, and priority in the affiliate program."),
					),
					h.A(
						h.Class("success-social"),
						h.Href("https://t.me/arkive_sh"),
						g.Attr("rel", "noreferrer"),
						g.Attr("target", "_blank"),
						components.Icon(components.IconProps{
							Name:       "telegram",
							Size:       "lg",
							Decorative: true,
						}),
						g.Text("Follow our telegram for more rewards and information."),
					),
				},
			}),
		),
	)
}
