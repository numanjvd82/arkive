package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	. "arkive/core/web"
)

func HomePage() g.Node {
	return Layout(LayoutData{
		Title: "Arkive",
		CSS:   "/static/pages/home.css",
	},
		h.Div(
			h.Class("page home"),
			homeHeader(),
			homeHero(),
			homeHighlights(),
			homeWorkflow(),
			homeSecurity(),
			homeCTA(),
			homeFooter(),
		),
	)
}

func homeHeader() g.Node {
	return h.Header(
		h.Class("site-header"),
		h.Div(
			h.Class("container nav"),
			h.Div(
				h.Class("logo-mark"),
				h.Span(h.Class("logo-dot")),
				h.Span(h.Class("logo-text"), g.Text("Arkive")),
			),
			h.Nav(
				h.Class("nav-links"),
				h.A(h.Href("#features"), g.Text("Features")),
				h.A(h.Href("#workflow"), g.Text("Workflow")),
				h.A(h.Href("#security"), g.Text("Security")),
				h.A(h.Href("#pricing"), g.Text("Pricing")),
			),
			h.Div(
				h.Class("nav-actions"),
				h.Button(
					h.Class("button secondary theme-toggle"),
					h.Type("button"),
					h.ID("theme-toggle"),
					g.Attr("aria-pressed", "true"),
					h.Span(h.Class("theme-icon"), g.Text("◐")),
					h.Span(h.Class("theme-label"), g.Text("Theme")),
				),
				h.A(h.Class("button secondary"), h.Href("/login"), g.Text("Login")),
				h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Create account")),
			),
		),
	)
}

func homeHero() g.Node {
	return h.Main(
		h.Class("hero"),
		h.Div(
			h.Class("container hero-grid"),
			h.Div(
				h.Class("hero-copy"),
				h.Span(h.Class("eyebrow"), g.Text("Your archive, always ready")),
				h.H1(g.Text("A calm place for every project file, note, and decision.")),
				h.P(g.Text("Collect research, drafts, and discussions in one space. Arkive helps you keep momentum while your team stays aligned.")),
				h.Div(
					h.Class("hero-actions"),
					h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Start free")),
					h.A(h.Class("button secondary"), h.Href("/login"), g.Text("View demo")),
				),
				h.Div(
					h.Class("hero-stats"),
					statItem("24/7", "instant access"),
					statItem("3x", "faster handoffs"),
					statItem("SOC2", "ready vault"),
				),
			),
			h.Div(
				h.Class("hero-panel"),
				h.Div(
					h.Class("hero-card"),
					h.P(h.Class("card-title"), g.Text("Latest capture")),
					h.Div(
						h.Class("card-body"),
						h.Div(h.Class("card-chip"), g.Text("Product research")),
						h.H3(g.Text("Competitive teardown")),
						h.P(g.Text("Saved 12 links, 4 notes, and 2 design drafts.")),
						h.Div(
							h.Class("card-meta"),
							h.Span(g.Text("Updated 2 hours ago")),
							h.Span(g.Text("6 collaborators")),
						),
					),
				),
				h.Div(
					h.Class("hero-card secondary"),
					h.P(h.Class("card-title"), g.Text("Sharing status")),
					h.Ul(
						h.Class("card-list"),
						h.Li(g.Text("Marketing: view access")),
						h.Li(g.Text("Product: edit access")),
						h.Li(g.Text("Design: edit access")),
					),
				),
			),
		),
	)
}

func homeHighlights() g.Node {
	return h.Section(
		h.ID("features"),
		h.Class("feature-section"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("section-heading"),
				h.H2(g.Text("Designed to keep your context alive.")),
				h.P(g.Text("Capture quickly, organize with clarity, and share without losing the thread.")),
			),
			h.Div(
				h.Class("feature-grid"),
				featureCard("Capture anything", "Save files, links, voice notes, and images in seconds from any device."),
				featureCard("Collections that breathe", "Structure your archive by project, milestone, or team without rigid folders."),
				featureCard("Smart reminders", "Get subtle nudges when a project is drifting or a file is outdated."),
				featureCard("Clean handoffs", "Deliver polished summaries for stakeholders with one click."),
			),
		),
	)
}

func homeWorkflow() g.Node {
	return h.Section(
		h.ID("workflow"),
		h.Class("workflow"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("section-heading"),
				h.H2(g.Text("A workflow that feels lighter.")),
				h.P(g.Text("From capture to delivery, Arkive keeps your team in flow without noisy tools.")),
			),
			h.Div(
				h.Class("step-grid"),
				stepCard("01", "Collect", "Drop in anything you find. Arkive tags and connects it automatically."),
				stepCard("02", "Curate", "Pin the key moments and draft summaries together."),
				stepCard("03", "Share", "Publish a living brief or export a clean report."),
			),
		),
	)
}

func homeCTA() g.Node {
	return h.Section(
		h.ID("pricing"),
		h.Class("cta"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("cta-card"),
				h.Div(
					h.Class("cta-copy"),
					h.Span(h.Class("eyebrow"), g.Text("Launch your Arkive")),
					h.H2(g.Text("Keep every project decision discoverable.")),
					h.P(g.Text("Start with a free workspace, invite your team, and upgrade when you are ready.")),
				),
				h.Div(
					h.Class("cta-actions"),
					h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Create workspace")),
					h.A(h.Class("button secondary"), h.Href("/login"), g.Text("Talk to sales")),
				),
			),
		),
	)
}

func homeSecurity() g.Node {
	return h.Section(
		h.ID("security"),
		h.Class("security"),
		h.Div(
			h.Class("container security-grid"),
			h.Div(
				h.Class("security-copy"),
				h.H2(g.Text("Secure by default, calm by design.")),
				h.P(g.Text("Your archive is encrypted, audited, and ready for compliance from day one.")),
				h.Div(
					h.Class("security-tags"),
					h.Span(h.Class("tag"), g.Text("SOC2 ready")),
					h.Span(h.Class("tag"), g.Text("Role based access")),
					h.Span(h.Class("tag"), g.Text("Private by default")),
				),
			),
			h.Div(
				h.Class("security-panel"),
				securityRow("Encryption at rest", "All files and notes are encrypted with modern standards."),
				securityRow("Activity logging", "See who accessed or shared any collection."),
				securityRow("Export control", "Invite partners with time-bound access."),
			),
		),
	)
}

func homeFooter() g.Node {
	return h.Footer(
		h.Class("site-footer"),
		h.Div(
			h.Class("container footer-grid"),
			h.Div(
				h.Class("footer-brand"),
				h.H3(g.Text("Arkive")),
				h.P(g.Text("Capture, curate, and share what matters.")),
			),
			h.Div(
				h.Class("footer-links"),
				h.A(h.Href("#features"), g.Text("Features")),
				h.A(h.Href("#workflow"), g.Text("Workflow")),
				h.A(h.Href("#security"), g.Text("Security")),
				h.A(h.Href("#pricing"), g.Text("Pricing")),
			),
			h.Div(
				h.Class("footer-links"),
				h.A(h.Href("#"), g.Text("Docs")),
				h.A(h.Href("#"), g.Text("Changelog")),
				h.A(h.Href("#"), g.Text("Support")),
			),
		),
	)
}

func featureCard(title, body string) g.Node {
	return h.Article(
		h.Class("feature-card"),
		h.H3(g.Text(title)),
		h.P(g.Text(body)),
	)
}

func statItem(value, label string) g.Node {
	return h.Div(
		h.Class("stat"),
		h.Span(h.Class("stat-value"), g.Text(value)),
		h.Span(h.Class("stat-label"), g.Text(label)),
	)
}

func stepCard(step, title, body string) g.Node {
	return h.Article(
		h.Class("step-card"),
		h.Span(h.Class("step-index"), g.Text(step)),
		h.H3(g.Text(title)),
		h.P(g.Text(body)),
	)
}

func securityRow(title, body string) g.Node {
	return h.Div(
		h.Class("security-row"),
		h.H3(g.Text(title)),
		h.P(g.Text(body)),
	)
}
