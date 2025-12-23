package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

func LoginPage() web.Page {
	return web.Page{
		Title: "Arkive · Login",
		CSS:   []string{"/web/pages/login.css"},
		Body:  loginBody(),
	}
}

func loginBody() g.Node {
	return h.Div(
		h.Class("auth-container"),
		h.Form(
			h.Class("auth-form"),
			g.Attr("method", "POST"),
			h.H2(g.Text("Login")),
			h.Label(
				g.Attr("for", "email"),
				g.Text("Email"),
			),
			h.Input(
				g.Attr("type", "email"),
				g.Attr("name", "email"),
				g.Attr("id", "email"),
				g.Attr("required", "required"),
			),
			h.Label(
				g.Attr("for", "password"),
				g.Text("Password"),
			),
			h.Input(
				g.Attr("type", "password"),
				g.Attr("name", "password"),
				g.Attr("id", "password"),
				g.Attr("required", "required"),
			),
			h.Button(
				g.Attr("type", "submit"),
				g.Text("Login"),
			),
		),
	)
}
