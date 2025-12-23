package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	. "arkive/core/web"
)

type AuthPageData struct {
	Title   string
	CSS     string
	Error   string
	Name    string
	Email   string
}

func LoginPage(data AuthPageData) g.Node {
	return Layout(LayoutData{
		Title: data.Title,
		CSS:   data.CSS,
	},
		h.Main(
			h.Class("auth-page"),
			h.Div(
				h.Class("auth-shell"),
				h.H1(g.Text("Login")),
				h.Form(
					h.Class("auth-card"),
					h.Method("post"),
					h.Action("/login"),
					inputField("Email", "email", "email", "you@company.com", data.Email),
					inputField("Password", "password", "password", "••••••••", ""),
					authError(data.Error),
					h.Button(h.Class("button primary"), h.Type("submit"), g.Text("Login")),
				),
				h.P(
					h.Class("auth-footer"),
					g.Text("New here? "),
					h.A(h.Href("/signup"), g.Text("Create an account")),
				),
			),
		),
	)
}

func SignupPage(data AuthPageData) g.Node {
	return Layout(LayoutData{
		Title: data.Title,
		CSS:   data.CSS,
	},
		h.Main(
			h.Class("auth-page"),
			h.Div(
				h.Class("auth-shell"),
				h.H1(g.Text("Sign Up")),
				h.Form(
					h.Class("auth-card"),
					h.Method("post"),
					h.Action("/signup"),
					inputField("Brand name", "brand_name", "text", "studio-arkive", data.Name),
					inputField("Email", "email", "email", "you@company.com", data.Email),
					inputField("Password", "password", "password", "Create a password", ""),
					authError(data.Error),
					h.Button(h.Class("button primary"), h.Type("submit"), g.Text("Create account")),
				),
				h.P(
					h.Class("auth-footer"),
					g.Text("Already have an account? "),
					h.A(h.Href("/login"), g.Text("Login")),
				),
			),
		),
	)
}

func inputField(label, name, inputType, placeholder, value string) g.Node {
	return h.Div(
		h.Class("auth-field"),
		h.Label(
			h.Class("auth-label"),
			g.Text(label),
		),
		h.Input(
			h.Type(inputType),
			h.Name(name),
			h.Placeholder(placeholder),
			h.Value(value),
			h.Required(),
		),
	)
}

func authError(message string) g.Node {
	if message == "" {
		return g.Group{}
	}
	return h.P(h.Class("auth-error"), g.Text(message))
}
