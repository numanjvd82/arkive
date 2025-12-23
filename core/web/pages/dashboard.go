package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

func DashboardPage() web.Page {
	return web.Page{
		Title: "Arkive · Dashboard",
		CSS:   []string{"/web/pages/dashboard.css"},
		Body: h.Main(
			h.Class("dashboard"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("dashboard-header"),
					h.Div(
						h.Class("dashboard-brand"),
						h.Span(h.Class("logo-dot")),
						h.Span(h.Class("logo-text"), g.Text("Arkive")),
					),
					h.Form(
						h.Method("post"),
						h.Action("/logout"),
						h.Button(h.Class("button secondary"), h.Type("submit"), g.Text("Logout")),
					),
				),
				h.Section(
					h.Class("dashboard-hero"),
					h.H1(g.Text("Welcome back.")),
					h.P(g.Text("Your workspace is ready. This is a simple placeholder dashboard for auth testing.")),
				),
				h.Div(
					h.Class("dashboard-grid"),
					dashboardCard("Collections", "3 active"),
					dashboardCard("Shared links", "12 saved"),
					dashboardCard("Last updated", "2 hours ago"),
				),
			),
		),
	}
}

func dashboardCard(title, body string) g.Node {
	return h.Div(
		h.Class("dashboard-card"),
		h.Span(h.Class("dashboard-label"), g.Text(title)),
		h.P(h.Class("dashboard-value"), g.Text(body)),
	)
}
