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
					g.Text("Arkive is free, with no hard storage limits. Upload, store, and share freely without worrying about space."),
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
								pricingRow("Price", "Free", "Contact us"),
								pricingRow("Storage", "Unlimited (personal use)", "Unlimited (higher tiers)"),
								pricingRow("Upload speed", "Up to 10 GB/month full-speed", "1 TB/month or more"),
								pricingRow("Share links", "Unlimited", "Unlimited"),
								pricingRow("Share creation rate", "6/min (burst 12)", "30/min (burst 60)"),
								pricingRow("Download/stream rate", "60/min (burst 120)", "300/min (burst 600)"),
								pricingRow("Concurrent uploads", "2 files at once", "Up to 10"),
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
						h.Li(g.Text("You can upload as much as you like within the monthly speed limits.")),
						h.Li(g.Text("Share as many links as you want, as long as it is real usage (not link farms or spam).")),
						h.Li(g.Text("No abusive behavior, including hotlinking, scraping, or using Arkive as a backup mirror.")),
						h.Li(g.Text("If you need more speed or higher limits, just ask.")),
					),
					h.P(
						h.Class("fairuse-note"),
						g.Text("We rarely throttle and will reach out if anything needs attention. Contact support@arkive.sh."),
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
