package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type PricingPageProps struct {
	Ctx PageContext
}

func PricingPage(props PricingPageProps) web.Page {
	_ = props
	return web.Page{
		Title: "Arkive · Pricing & Fair Use",
		CSS:   []string{"/web/pages/pricing.css"},
		JS:    []string{"/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		Body:  pricingBody(),
	}
}

func pricingBody() g.Node {
	return h.Div(
		h.Class("pricing-page"),
		h.Div(
			h.Class("pricing-hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Pricing & Fair Use")),
				h.P(h.Class("pricing-meta"), g.Text("Last updated January 02, 2026")),
				h.P(
					h.Class("pricing-intro"),
					g.Text("Arkive is free with generous storage and fair-use limits. Unlimited uploads, fair-use storage, and unlimited retention while active."),
				),
			),
		),
		h.Div(
			h.Class("pricing-content"),
			h.Div(
				h.Class("container"),
				h.Section(
					h.Class("pricing-card"),
					h.Div(
						h.Class("pricing-table-wrap"),
						h.Table(
							h.Class("pricing-table"),
							h.THead(
								h.Tr(
									h.Th(g.Text("Feature")),
									h.Th(g.Text("Free")),
									h.Th(g.Text("Premium (coming soon)")),
								),
							),
							h.TBody(
								pricingRow("Price", "Free", "Coming soon"),
								pricingRow("Storage", "10 GB included (Fair Use)", "Unlimited (higher tiers)"),
								pricingRow("Retention", "Unlimited while active", "Unlimited"),
								pricingRow("Upload speed", "Normal", "Priority"),
								pricingRow("File limit", "10,000 files", "Higher limits"),
								pricingRow("Share links", "Unlimited", "Unlimited"),
								pricingRow("Ads", "Yes", "No"),
								pricingRow("Support", "Standard email support", "Priority support"),
							),
						),
					),
				),
				h.Section(
					h.Class("pricing-card fairuse-card"),
					h.H2(g.Text("Fair Use Policy")),
					h.P(
						h.Class("fairuse-intro"),
						g.Text("Arkive is built for real people, not bots or servers. To keep things fast and fair, we use a few automatic limits."),
					),
					h.Ul(
						h.Class("fairuse-list"),
						h.Li(g.Text("Uploads are unlimited. Storage is capped at 10 GB total for free accounts.")),
						h.Li(g.Text("Accounts with no login or file activity for 30 days may be archived.")),
						h.Li(g.Text("Activity means owner logins or authenticated actions; public link views do not count.")),
						h.Li(g.Text("Free restores are limited to 2 GB/day.")),
						h.Li(g.Text("Archived files are not immediately accessible. Log in to restore them (may take a short time).")),
						h.Li(g.Text("Archived files are deleted after 7 days unless they are restored or upgraded.")),
						h.Li(g.Text("We email you before archiving or deleting files.")),
						h.Li(g.Text("Share as many links as you want, as long as it is real usage (not link farms or spam).")),
						h.Li(g.Text("No abusive behavior, including hotlinking, scraping, or using Arkive as a backup mirror.")),
						h.Li(g.Text("Public sharing of illegal content will result in immediate suspension.")),
						h.Li(g.Text("If you need more speed or higher limits, just ask.")),
					),
					h.H3(g.Text("What happens if limits are hit")),
					h.Ul(
						h.Class("fairuse-list"),
						h.Li(g.Text("Temporary throttling or reduced speeds.")),
						h.Li(g.Text("Limits reduction or archive state for inactive accounts.")),
						h.Li(g.Text("Deletion after notice for prolonged inactivity.")),
						h.Li(g.Text("Account suspension for illegal or abusive use.")),
					),
					h.P(
						h.Class("fairuse-note"),
						g.Text("We rarely need to enforce limits and will reach out if anything needs attention. Contact support@arkive.sh."),
					),
					h.P(
						h.Class("fairuse-intro"),
						g.Text("Have a special use case (YouTuber? Educator? Community archive?) We would love to help. Just reach out."),
					),
				),
			),
		),
	)
}

func pricingRow(feature, free, premium string) g.Node {
	return h.Tr(
		h.Th(g.Text(feature)),
		h.Td(g.Text(free)),
		h.Td(g.Text(premium)),
	)
}
