package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type PrivacyPageProps struct {
	Ctx PageContext
}

func PrivacyPage(props PrivacyPageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "Arkive · Privacy Policy",
		Description:   "Read Arkive's privacy policy to understand how we collect, use, and protect your information.",
		CanonicalPath: "/privacy",
		OGImage:       DefaultOGImage,
		Robots:        RobotsIndex,
		CSS:           []string{"/web/pages/privacy.css"},
		Body:          privacyBody(),
	}
}

func privacyBody() g.Node {
	return h.Div(
		h.Class("privacy-page"),
		h.Div(
			h.Class("privacy-hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Privacy Policy")),
				h.P(h.Class("privacy-meta"), g.Text("Last updated January 02, 2026")),
				h.P(
					h.Class("privacy-intro"),
					g.Text("This policy explains how Arkive (arkive.sh) collects, uses, and shares personal information when you use our website, web app, and mobile app (when available)."),
				),
				h.P(
					h.Class("privacy-intro"),
					g.Text("Related documents: "),
					h.A(h.Href("/terms"), h.Class("text-link"), g.Text("Terms of Service")),
					g.Text(", "),
					h.A(h.Href("/cookies"), h.Class("text-link"), g.Text("Cookie Policy")),
					g.Text(", and "),
					h.A(h.Href("/aup"), h.Class("text-link"), g.Text("Acceptable Use Policy")),
					g.Text(", and "),
					h.A(h.Href("/abuse"), h.Class("text-link"), g.Text("Copyright & Abuse Reporting")),
					g.Text("."),
				),
			),
		),
		h.Div(
			h.Class("privacy-content"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("privacy-card"),
					privacySection(
						"1. Information we collect",
						g.Group([]g.Node{
							h.P(g.Text("Information you provide: account details like name, username, and email; credentials; files and metadata you upload; sharing settings; abuse or copyright reports; and messages you send to support.")),
							h.P(g.Text("Information from third parties: if you sign in with Google, we receive your name, email, profile photo, and Google account ID. In the Android app, purchases are processed by Google Play, which provides us with purchase status and receipts. Rewarded ad partners may provide ad interaction details needed to grant rewards.")),
							h.P(g.Text("Information collected automatically: IP address, device and browser details, usage and log data, and security signals. We use essential cookies to keep you signed in and protect sessions.")),
						}),
					),
					privacySection(
						"2. How we use information",
						g.Group([]g.Node{
							h.P(g.Text("We use information to create and manage accounts, process uploads and sharing, deliver files, provide support, detect abuse, maintain security, and improve performance.")),
							h.P(g.Text("We also use limited analytics to understand traffic patterns and service reliability.")),
						}),
					),
					privacySection(
						"3. Ads, analytics, and rewarded offers",
						g.Group([]g.Node{
							h.P(g.Text("We show ads on the website using a third-party ad provider. We do not use targeted advertising and do not sell or share personal information for targeted ads.")),
							h.P(g.Text("We use Cloudflare Web Analytics to understand aggregate traffic and performance. Cloudflare may process device and log data on our behalf.")),
							h.P(g.Text("On the website and in the mobile app, you may choose to watch rewarded ads to receive temporary storage boosts or rate limit increases. To award rewards and prevent fraud, our ad partners may receive a user identifier and basic device or ad interaction data.")),
							h.P(g.Text("Rewards are promotional, subject to availability and fraud checks, and may change or end at any time.")),
							h.P(g.Text("We may display plan information on the website, but purchases are completed in the Android app through Google Play.")),
						}),
					),
					privacySection(
						"4. How we share information",
						g.Group([]g.Node{
							h.P(g.Text("We share information with service providers who help us operate the service, such as hosting, storage, analytics, and email delivery. These providers are bound by contracts to protect your data.")),
							h.P(g.Text("We share information with Google when you use Google sign-in, and with Google Play when you purchase a plan in the Android app.")),
							h.P(g.Text("We may share information to comply with legal obligations, protect our users, and enforce our terms.")),
							h.P(g.Text("If Arkive is involved in a merger, acquisition, or asset sale, information may be transferred as part of that transaction.")),
						}),
					),
					privacySection(
						"5. Data retention",
						g.Group([]g.Node{
							h.P(g.Text("We keep account and file data for as long as your account is active. Accounts with no login or file activity for 30 days may be archived, and archived files are deleted after 7 days unless they are restored or upgraded.")),
							h.P(g.Text("We may notify you by email before archiving or deleting files.")),
							h.P(g.Text("We delete or anonymize data when it is no longer needed, unless we are required to keep it for legal or security reasons.")),
						}),
					),
					privacySection(
						"6. Your choices and rights (worldwide)",
						g.Group([]g.Node{
							h.P(g.Text("You can update account details in your settings. You can also request access, correction, or deletion of your data by emailing support@arkive.sh.")),
							h.P(g.Text("Depending on your location, you may have additional rights, such as access, correction, deletion, portability, restriction, objection, or withdrawal of consent.")),
							h.P(g.Text("EEA, UK, and Switzerland: you may object to processing, request restriction, or request data portability, and you can complain to your local data protection authority.")),
							h.P(g.Text("United States: depending on your state, you may request access, deletion, correction, or opt out of certain processing. You may also use an authorized agent to submit a request on your behalf, subject to verification.")),
							h.P(g.Text("Canada: you may request access or correction and may withdraw consent where applicable.")),
							h.P(g.Text("Australia and New Zealand: you may request access or correction and can file a complaint with the OAIC or the NZ Privacy Commissioner.")),
							h.P(g.Text("South Africa: you may request access or correction and can file a complaint with the Information Regulator.")),
							h.P(g.Text("Brazil: you may request access, correction, anonymization, deletion, portability, or information about data sharing under the LGPD.")),
						}),
					),
					privacySection(
						"7. Legal bases for processing (EEA, UK, Switzerland)",
						g.Group([]g.Node{
							h.P(g.Text("If you are in the EEA, UK, or Switzerland, we rely on one or more of the following legal bases: performance of a contract, consent, compliance with legal obligations, and our legitimate interests (such as fraud prevention, security, and service improvement).")),
						}),
					),
					privacySection(
						"8. International data transfers",
						g.Group([]g.Node{
							h.P(g.Text("Arkive operates globally. Your information may be processed in countries other than where you live. Where required by law, we use appropriate safeguards for cross-border transfers.")),
						}),
					),
					privacySection(
						"9. Security",
						g.Group([]g.Node{
							h.P(g.Text("We use technical and organizational safeguards to protect your information. No system can be guaranteed to be 100 percent secure, so please use the service responsibly.")),
						}),
					),
					privacySection(
						"10. Children",
						g.Group([]g.Node{
							h.P(g.Text("Arkive is not intended for children under 13. If you believe a child has provided personal information, contact us and we will take appropriate steps.")),
						}),
					),
					privacySection(
						"11. Contact",
						g.Group([]g.Node{
							h.P(
								g.Text("Questions or requests? Email "),
								h.A(h.Href("mailto:support@arkive.sh"), h.Class("text-link"), g.Text("support@arkive.sh")),
								g.Text("."),
							),
						}),
					),
				),
			),
		),
	)
}

func privacySection(title string, content g.Node) g.Node {
	return h.Section(
		h.Class("privacy-section"),
		h.H2(g.Text(title)),
		content,
	)
}
