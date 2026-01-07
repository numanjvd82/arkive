package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type ContactPageProps struct {
	Ctx PageContext
}

func ContactPage(props ContactPageProps) web.Page {
	_ = props
	return web.Page{
		Title: "Arkive · Contact",
		CSS:   []string{"/web/pages/contact.css"},
		JS:    []string{"/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		Body:  contactBody(),
	}
}

func contactBody() g.Node {
	return h.Div(
		h.Class("contact-page"),
		components.InlineStyle(components.InputCSS),
		contactHero(),
		contactContent(),
	)
}

func contactHero() g.Node {
	return h.Section(
		h.Class("contact-hero"),
		h.Div(
			h.Class("container contact-hero-grid"),
			h.Div(
				h.Class("contact-hero-copy"),
				h.Span(h.Class("eyebrow"), g.Text("Contact")),
				h.H1(g.Text("Let us know how we can help you share.")),
				h.P(g.Text("Reach out for support, partnerships, or anything Arkive. We respond quickly and keep every conversation human.")),
				h.Div(
					h.Class("contact-highlights"),
					highlightCard("24h response", "Most support requests are answered within a day."),
					highlightCard("Security-first", "We keep your data private and protected."),
				),
			),
			h.Div(
				h.Class("contact-hero-card"),
				h.Span(h.Class("contact-pill"), g.Text("Availability")),
				h.H3(g.Text("Support hours")),
				h.P(g.Text("Monday to Friday, 9am–6pm (GMT +2).")),
				h.Div(
					h.Class("contact-hero-meta"),
					h.Div(
						h.Span(h.Class("meta-label"), g.Text("Priority")),
						h.Span(h.Class("meta-value"), g.Text("Security & billing")),
					),
					h.Div(
						h.Span(h.Class("meta-label"), g.Text("Status")),
						h.Span(h.Class("meta-value"), g.Text("All systems normal")),
					),
				),
			),
		),
	)
}

func contactContent() g.Node {
	return h.Section(
		h.Class("contact-content"),
		h.Div(
			h.Class("container contact-grid"),
			h.Form(
				h.Class("contact-form"),
				h.Action("mailto:support@arkive.sh"),
				h.Method("post"),
				h.EncType("text/plain"),
				h.H2(g.Text("Send a message")),
				h.P(
					h.Class("contact-lead"),
					g.Text("Tell us what you are working on and we will get back to you with next steps."),
				),
				h.Div(
					h.Class("form-row"),
					formInput("contact-name", "Full name", "name", "name", "text", "Alex Doe", true),
					formInput("contact-email", "Email", "email", "email", "email", "you@company.com", true),
				),
				h.Div(
					h.Class("form-row"),
					formInput("contact-company", "Company", "company", "organization", "text", "Studio or team", false),
					formSelect(),
				),
				h.Div(
					h.Class("form-field"),
					h.Label(
						h.Class("form-label"),
						g.Attr("for", "contact-message"),
						g.Text("Message"),
					),
					h.Textarea(
						h.Class("form-input contact-textarea"),
						h.ID("contact-message"),
						h.Name("message"),
						h.Placeholder("Share a few details so we can route you to the right person."),
						g.Attr("rows", "6"),
						g.Attr("required", "required"),
					),
				),
				h.Div(
					h.Class("contact-actions"),
					h.Button(
						h.Class("button primary"),
						h.Type("submit"),
						g.Text("Send message"),
					),
					h.A(
						h.Class("button secondary"),
						h.Href("mailto:support@arkive.sh"),
						g.Text("Email directly"),
					),
				),
				h.P(
					h.Class("contact-note"),
					g.Text("Prefer something secure? Email security@arkive.sh for sensitive reports."),
				),
			),
			h.Aside(
				h.Class("contact-aside"),
				contactInfoCard(
					"Direct lines",
					[]g.Node{
						contactInfoRow("Support", "support@arkive.sh"),
						contactInfoRow("Security", "security@arkive.sh"),
					},
				),
				contactInfoCard(
					"Community",
					[]g.Node{
						h.P(g.Text("Join our community updates and releases.")),
						h.A(h.Class("text-link"), h.Href("https://t.me/arkive"), g.Text("t.me/arkive")),
					},
				),
				contactInfoCard(
					"Quick answers",
					[]g.Node{
						h.P(g.Text("Looking for pricing or policies? We keep them transparent.")),
						h.Div(
							h.Class("contact-links"),
							h.A(h.Class("text-link"), h.Href("/pricing"), g.Text("Pricing")),
							h.A(h.Class("text-link"), h.Href("/privacy"), g.Text("Privacy")),
							h.A(h.Class("text-link"), h.Href("/aup"), g.Text("Acceptable Use")),
							h.A(h.Class("text-link"), h.Href("/abuse"), g.Text("Copyright & Abuse")),
						),
					},
				),
			),
		),
	)
}

func highlightCard(title, body string) g.Node {
	return h.Div(
		h.Class("highlight-card"),
		h.H4(g.Text(title)),
		h.P(g.Text(body)),
	)
}

func formInput(id, label, name, autocomplete, inputType, placeholder string, required bool) g.Node {
	if inputType == "" {
		inputType = "text"
	}
	return h.Div(
		h.Class("form-field"),
		h.Label(
			h.Class("form-label"),
			g.Attr("for", id),
			g.Text(label),
		),
		h.Input(
			h.Class("form-input"),
			h.ID(id),
			h.Name(name),
			g.Attr("type", inputType),
			g.Attr("autocomplete", autocomplete),
			h.Placeholder(placeholder),
			g.If(required, g.Attr("required", "required")),
		),
	)
}

func formSelect() g.Node {
	return h.Div(
		h.Class("form-field"),
		h.Label(
			h.Class("form-label"),
			g.Attr("for", "contact-topic"),
			g.Text("Topic"),
		),
		h.Select(
			h.Class("form-input"),
			h.ID("contact-topic"),
			h.Name("topic"),
			h.Option(h.Value("support"), g.Text("Product support")),
			h.Option(h.Value("billing"), g.Text("Billing question")),
			h.Option(h.Value("partnerships"), g.Text("Partnerships")),
			h.Option(h.Value("press"), g.Text("Press or media")),
			h.Option(h.Value("other"), g.Text("Something else")),
		),
	)
}

func contactInfoCard(title string, body []g.Node) g.Node {
	return h.Div(
		h.Class("contact-info-card"),
		h.H3(g.Text(title)),
		g.Group(body),
	)
}

func contactInfoRow(label, value string) g.Node {
	return h.Div(
		h.Class("contact-info-row"),
		h.Span(h.Class("contact-info-label"), g.Text(label)),
		h.A(h.Class("text-link"), h.Href("mailto:"+value), g.Text(value)),
	)
}
