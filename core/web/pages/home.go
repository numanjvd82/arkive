package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type HomePageProps struct {
	Ctx PageContext
}

func HomePage(props HomePageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "Arkive · Share with freedom",
		Description:   "Fast, secure file sharing with zero friction. Share links instantly, add passwords, and let anyone download without an account.",
		CanonicalPath: "/",
		OGImage:       DefaultOGImage,
		Robots:        RobotsIndex,
		JSONLD:        `{"@context":"https://schema.org","@type":"WebSite","name":"Arkive","url":"https://arkive.sh","description":"Fast, secure file sharing with zero friction."}`,
		CSS:           []string{"/web/pages/home.css"},
		HideNav:       true,
		Body: h.Div(
			h.Class("page home"),
			marketingHeader(),
			homeHero(),
			homeFeatures(),
			homeSharing(),
			homeSecurity(),
			homeComparison(),
			homeRoadmap(),
			homeCTA(),
			marketingFooter(),
		),
	}
}

func homeHero() g.Node {
	return h.Main(
		h.Class("hero"),
		h.Div(
			h.Class("container hero-grid"),
			h.Div(
				h.Class("hero-copy"),
				h.Span(h.Class("eyebrow"), g.Text("Share with freedom")),
				h.H1(g.Text("Fast, secure file sharing with zero friction.")),
				h.P(g.Text("Arkive gives you generous storage with unlimited retention while active. Share links instantly, add passwords, and let anyone download—no account required.")),
				h.P(
					h.Class("hero-note"),
					g.Text("Fair use applies. "),
					h.A(h.Href("/pricing"), h.Class("text-link"), g.Text("See pricing")),
					g.Text("."),
				),
				h.Div(
					h.Class("hero-actions"),
					h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Start sharing")),
					h.A(h.Class("button secondary"), h.Href("#features"), g.Text("See features")),
				),
				h.Div(
					h.Class("hero-stats"),
					statItem("10 GB", "included"),
					statItem("Active", "retention"),
					statItem("No login", "for shared links"),
				),
			),
			h.Div(
				h.Class("hero-panel"),
				h.Div(
					h.Class("hero-card"),
					h.P(h.Class("card-title"), g.Text("Share link")),
					h.H3(g.Text("arkive.sh/s/quiet-sunrise")),
					h.P(g.Text("Password protected · Unlimited downloads")),
					h.Div(
						h.Class("card-list"),
						h.Span(g.Text("• Instant uploads")),
						h.Span(g.Text("• Link previews")),
						h.Span(g.Text("• Zero sign-in")),
					),
				),
				h.Div(
					h.Class("hero-card secondary"),
					h.P(h.Class("card-title"), g.Text("Freedom pack")),
					h.Div(
						h.Class("freedom-grid"),
						freedomItem("Lightning fast", "Optimized for quick uploads and downloads."),
						freedomItem("Secure sharing", "Encrypted storage and optional passwords."),
						freedomItem("Open source", "Public roadmap and code coming soon."),
					),
				),
			),
		),
	)
}

func homeFeatures() g.Node {
	return h.Section(
		h.ID("features"),
		h.Class("feature-section"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("section-heading"),
				h.H2(g.Text("Everything you need to share without friction.")),
				h.P(g.Text("Built for creators, teams, and anyone who wants to move files fast without hoops.")),
			),
			h.Div(
				h.Class("feature-grid"),
				featureCard("Lightning fast", "Uploads and downloads tuned for speed across the globe."),
				featureCard("No accounts for viewers", "Anyone can access shared files without logging in."),
				featureCard("Password protected", "Lock shared links with a password whenever you want."),
				featureCard("Retention while active", "Files stay available while your account is active."),
				featureCard("Secure by default", "Encrypted storage and privacy-first defaults."),
				featureCard("Open source soon", "We are preparing a public repo and roadmap."),
			),
		),
	)
}

func homeSharing() g.Node {
	return h.Section(
		h.ID("sharing"),
		h.Class("sharing"),
		h.Div(
			h.Class("container sharing-grid"),
			h.Div(
				h.Class("sharing-copy"),
				h.H2(g.Text("Share in seconds, keep control forever.")),
				h.P(g.Text("Upload any file, generate a link instantly, and set a password if needed. Recipients can download without creating an account.")),
				h.Div(
					h.Class("share-steps"),
					stepCard("01", "Upload", "Drop files or folders in one click."),
					stepCard("02", "Secure", "Add a password or keep it open."),
					stepCard("03", "Send", "Share the link anywhere, instantly."),
				),
			),
			h.Div(
				h.Class("sharing-panel"),
				h.Div(
					h.Class("panel-card"),
					h.P(h.Class("card-title"), g.Text("Preview")),
					h.H3(g.Text("Campaign assets.zip")),
					h.P(g.Text("Ready to download")),
					h.Div(
						h.Class("panel-tags"),
						h.Span(h.Class("tag"), g.Text("Password on")),
						h.Span(h.Class("tag"), g.Text("Active retention")),
						h.Span(h.Class("tag"), g.Text("Unlimited downloads")),
					),
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
				h.H2(g.Text("Secure, private, and built to last.")),
				h.P(g.Text("Arkive keeps your files protected with encryption and privacy-first defaults. Retention is unlimited while active.")),
				h.Div(
					h.Class("security-tags"),
					h.Span(h.Class("tag"), g.Text("Encrypted storage")),
					h.Span(h.Class("tag"), g.Text("Password locks")),
					h.Span(h.Class("tag"), g.Text("Active retention")),
				),
			),
			h.Div(
				h.Class("security-panel"),
				securityRow("Zero friction access", "Recipients never need to create an account."),
				securityRow("Fast global delivery", "Share links optimized for rapid downloads."),
				securityRow("Open source soon", "A transparent, community-driven roadmap."),
			),
		),
	)
}

func homeRoadmap() g.Node {
	return h.Section(
		h.ID("roadmap"),
		h.Class("roadmap"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("section-heading"),
				h.H2(g.Text("More freedom is coming.")),
				h.P(g.Text("We are building the next layer of Arkive with tools for creators, teams, and partners.")),
			),
			h.Div(
				h.Class("roadmap-grid"),
				roadmapCard("Affiliate program", "Earn well for every creator you bring."),
				roadmapCard("Mobile apps", "Android and iOS apps with instant sharing."),
				roadmapCard("Premium membership", "Advanced controls and priority support."),
				roadmapCard("Open source", "Public repo and community contributions."),
			),
		),
	)
}

func homeComparison() g.Node {
	return h.Section(
		h.Class("comparison"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("section-heading"),
				h.H2(g.Text("Arkive vs the usual options.")),
				h.P(g.Text("A quick look at how Arkive compares for modern sharing workflows.")),
			),
			h.Div(
				h.Class("comparison-table"),
				h.Div(
					h.Class("comparison-row comparison-head"),
					h.Span(g.Text("Feature")),
					h.Span(g.Text("Arkive")),
					h.Span(g.Text("Drive/Dropbox")),
					h.Span(g.Text("WeTransfer")),
				),
				comparisonRow("No login for viewers", "Yes", "No", "Yes"),
				comparisonRow("Password protection", "Yes", "Limited", "Paid"),
				comparisonRow("Retention while active", "Yes", "Varies", "Expires"),
				comparisonRow("Drop Pages (Share Pages)", "Coming soon", "No", "No"),
				comparisonRow("Share link updates", "Yes", "Limited", "No"),
			),
		),
	)
}

func homeCTA() g.Node {
	return h.Section(
		h.Class("cta"),
		h.Div(
			h.Class("container"),
			h.Div(
				h.Class("cta-card"),
				h.Div(
					h.Class("cta-copy"),
					h.Span(h.Class("eyebrow"), g.Text("Start free")),
					h.H2(g.Text("Generous storage today. Share instantly.")),
					h.P(g.Text("Create your account and start sharing instantly. Unlimited retention while active, no friction for recipients.")),
				),
				h.Div(
					h.Class("cta-actions"),
					h.A(h.Class("button primary"), h.Href("/signup"), g.Text("Create account")),
					h.A(h.Class("button secondary"), h.Href("/login"), g.Text("Log in")),
				),
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

func roadmapCard(title, body string) g.Node {
	return h.Article(
		h.Class("roadmap-card"),
		h.H3(g.Text(title)),
		h.P(g.Text(body)),
	)
}

func freedomItem(title, body string) g.Node {
	return h.Div(
		h.Class("freedom-item"),
		h.H4(g.Text(title)),
		h.P(g.Text(body)),
	)
}

func comparisonRow(feature, arkive, drive, wetransfer string) g.Node {
	return h.Div(
		h.Class("comparison-row"),
		h.Span(h.Class("comparison-feature"), g.Text(feature)),
		h.Span(h.Class("comparison-value"), g.Text(arkive)),
		h.Span(h.Class("comparison-value"), g.Text(drive)),
		h.Span(h.Class("comparison-value"), g.Text(wetransfer)),
	)
}
