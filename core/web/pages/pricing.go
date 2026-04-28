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
		Title:         "Arkive · Pricing & Fair Use",
		Description:   "See Arkive pricing and fair use policy for our storage-first file sharing MVP.",
		CanonicalPath: "/pricing",
		OGImage:       DefaultOGImage,
		Robots:        RobotsIndex,
		CSS:           []string{"/web/pages/pricing.css"},
		JS:            []string{"/static/monetag-onclick.js", "/static/monetag-vignette.js"},
		Body:          pricingBody(),
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
				h.P(h.Class("pricing-meta"), g.Text("Last updated April 28, 2026")),
				h.P(
					h.Class("pricing-intro"),
					g.Text("Arkive is a storage + file sharing MVP. Plans include storage caps and monthly bandwidth limits to keep the service stable and fair."),
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
									h.Th(g.Text("Starter")),
									h.Th(g.Text("Pro")),
									h.Th(g.Text("Max")),
								),
							),
							h.TBody(
								pricingRow5("Price", "Free", "$9/month", "$19/month", "$29/month"),
								pricingRow5("Storage", "10 GB", "100 GB", "500 GB", "1 TB"),
								pricingRow5("Bandwidth / month", "100 GB (Fair Use)", "250 GB", "750 GB", "1–2 TB (Fair Use)"),
								pricingRow5("Public share links", "Yes", "Yes", "Yes", "Yes"),
								pricingRow5("Ads", "Enabled", "No ads", "No ads", "No ads"),
								pricingRow5("Inactivity archive", "After 30 days", "After 30 days", "After 30 days", "After 30 days"),
								pricingRow5("Upload priority", "Standard", "Faster uploads", "Priority uploads", "Highest priority"),
								pricingRow5("Support", "Standard", "Standard", "Priority", "Highest priority"),
							),
						),
					),
				),
				h.Section(
					h.Class("pricing-card fairuse-card"),
					h.H2(g.Text("Fair Use Policy")),
					h.P(
						h.Class("fairuse-intro"),
						g.Text("Arkive is a storage and file sharing platform (MVP stage). Fair use limits help prevent abuse and keep performance consistent for everyone."),
					),
					h.Ul(
						h.Class("fairuse-list"),
						h.Li(g.Text("Storage limits are set per plan (see table above).")),
						h.Li(g.Text("Bandwidth limits apply per plan and reset monthly.")),
						h.Li(g.Text("Accounts with no login or file activity for 30 days may be archived.")),
						h.Li(g.Text("Activity means owner logins or authenticated actions; public link views do not count.")),
						h.Li(g.Text("Restores from archive are limited to 2 GB/day.")),
						h.Li(g.Text("Archived files may take a short time to restore.")),
						h.Li(g.Text("Archived files are deleted after 7 days unless restored or upgraded.")),
						h.Li(g.Text("We email you before archiving or deleting files.")),
						h.Li(g.Text("Abuse protection: no spam, link farms, scraping, hotlinking misuse, or using Arkive as a public mirror.")),
						h.Li(g.Text("Illegal content or repeated abuse can result in suspension.")),
					),
					h.H3(g.Text("What happens if limits are hit")),
					h.Ul(
						h.Class("fairuse-list"),
						h.Li(g.Text("Uploads/downloads may be throttled temporarily.")),
						h.Li(g.Text("Uploads may be paused until the next billing month or until you upgrade.")),
						h.Li(g.Text("Inactive accounts may be archived after notice.")),
						h.Li(g.Text("Severe or illegal abuse can result in suspension.")),
					),
					h.P(
						h.Class("fairuse-note"),
						g.Text("If you hit limits for legitimate reasons, email support@arkive.sh and we’ll help you pick the right plan or a higher limit."),
					),
				),
			),
		),
	)
}

func pricingRow5(feature, free, starter, pro, max string) g.Node {
	return h.Tr(
		h.Th(g.Text(feature)),
		h.Td(g.Text(free)),
		h.Td(g.Text(starter)),
		h.Td(g.Text(pro)),
		h.Td(g.Text(max)),
	)
}
