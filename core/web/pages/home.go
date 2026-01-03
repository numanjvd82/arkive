package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
)

type HomePageProps struct {
	Ctx PageContext
}

func HomePage(props HomePageProps) web.Page {
	_ = props
	return web.Page{
		Title:   "Arkive · Share with freedom",
		CSS:     []string{"/web/pages/home.css"},
		HideNav: true,
		Body: h.Div(
			h.Class("page home"),
			homeHeader(),
			homeHero(),
			homeFeatures(),
			homeSharing(),
			homeSecurity(),
			homeRoadmap(),
			homeCTA(),
			homeFooter(),
		),
	}
}

func homeHeader() g.Node {
	return h.Header(
		h.Class("site-header"),
		h.Div(
			h.Class("container nav"),
			components.BrandLogo(components.BrandLogoProps{
				Href:  "/",
				Class: "nav-brand",
			}),
			h.Nav(
				h.Class("nav-links"),
				h.A(h.Href("/pricing"), g.Text("Pricing")),
				h.A(h.Href("/contact"), g.Text("Contact")),
				h.A(h.Href("#features"), g.Text("Features")),
				h.A(h.Href("#sharing"), g.Text("Sharing")),
				h.A(h.Href("#security"), g.Text("Security")),
				h.A(h.Href("#roadmap"), g.Text("Roadmap")),
			),
			h.Div(
				h.Class("nav-actions"),
				h.A(h.Class("button secondary"), h.Href("/login"), g.Text("Log in")),
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
				h.Span(h.Class("eyebrow"), g.Text("Share with freedom")),
				h.H1(g.Text("Fast, secure file sharing with zero friction.")),
				h.P(g.Text("Arkive gives you unlimited storage with unlimited retention. Share links instantly, add passwords, and let anyone download—no account required.")),
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
					statItem("Unlimited", "storage"),
					statItem("0", "expiry limits"),
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
				featureCard("Unlimited retention", "No expiry dates. Your files stay available."),
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
						h.Span(h.Class("tag"), g.Text("No expiry")),
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
				h.P(g.Text("Arkive keeps your files protected with encryption and privacy-first defaults. No retention timers. No surprise deletions.")),
				h.Div(
					h.Class("security-tags"),
					h.Span(h.Class("tag"), g.Text("Encrypted storage")),
					h.Span(h.Class("tag"), g.Text("Password locks")),
					h.Span(h.Class("tag"), g.Text("Unlimited retention")),
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
					h.H2(g.Text("Unlimited storage today. Share instantly.")),
					h.P(g.Text("Create your account and start sharing instantly. No limits on retention, no friction for recipients.")),
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

func homeFooter() g.Node {
	return h.Footer(
		h.Class("site-footer"),
		h.Div(
			h.Class("container footer-grid"),
			h.Div(
				h.Class("footer-brand"),
				h.H3(g.Text("Arkive")),
				h.P(g.Text("Share files with freedom, speed, and security.")),
			),
			h.Div(
				h.Class("footer-links"),
				h.A(h.Href("/pricing"), g.Text("Pricing")),
				h.A(h.Href("#features"), g.Text("Features")),
				h.A(h.Href("#sharing"), g.Text("Sharing")),
				h.A(h.Href("#security"), g.Text("Security")),
				h.A(h.Href("#roadmap"), g.Text("Roadmap")),
			),
			h.Div(
				h.Class("footer-links"),
				h.A(h.Href("/login"), g.Text("Login")),
				h.A(h.Href("/signup"), g.Text("Create account")),
				h.A(h.Href("/contact"), g.Text("Contact")),
				h.A(h.Href("/privacy"), h.Class("text-link"), g.Text("Privacy Policy")),
				h.A(h.Href("/cookies"), h.Class("text-link"), g.Text("Cookie Policy")),
				h.A(h.Href("/terms"), h.Class("text-link"), g.Text("Terms of Service")),
				h.A(h.Href("/aup"), h.Class("text-link"), g.Text("Acceptable Use")),
				h.A(h.Href("/abuse"), h.Class("text-link"), g.Text("Copyright & Abuse")),
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
