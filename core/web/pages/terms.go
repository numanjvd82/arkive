package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
)

type TermsPageProps struct {
	Ctx PageContext
}

func TermsPage(props TermsPageProps) web.Page {
	_ = props
	return web.Page{
		Title:         "Arkive · Terms of Service",
		Description:   "Read the Arkive terms of service covering account use, content rules, and legal terms.",
		CanonicalPath: "/terms",
		OGImage:       DefaultOGImage,
		Robots:        RobotsIndex,
		CSS:           []string{"/web/pages/terms.css"},
		Body:          termsBody(),
	}
}

func termsBody() g.Node {
	return h.Div(
		h.Class("terms-page"),
		h.Div(
			h.Class("terms-hero"),
			h.Div(
				h.Class("container"),
				h.H1(g.Text("Terms of Service")),
				h.P(h.Class("terms-meta"), g.Text("Last updated January 02, 2026")),
				h.P(
					h.Class("terms-intro"),
					g.Text("These Terms of Service govern your use of Arkive (arkive.sh). By using the service, you agree to these terms, our Privacy Policy, and our Cookie Policy."),
				),
				h.P(
					h.Class("terms-intro"),
					g.Text("Related documents: "),
					h.A(h.Href("/privacy"), h.Class("text-link"), g.Text("Privacy Policy")),
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
			h.Class("terms-content"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("terms-card"),
					termsSection(
						"1. Eligibility",
						g.Group([]g.Node{
							h.P(g.Text("You must be at least 13 years old to use Arkive. If you are under the age of majority in your jurisdiction, you must have permission from a parent or legal guardian.")),
						}),
					),
					termsSection(
						"2. Accounts and security",
						g.Group([]g.Node{
							h.P(g.Text("You are responsible for your account and for keeping your credentials secure. Notify us immediately if you suspect unauthorized access.")),
						}),
					),
					termsSection(
						"3. Your content",
						g.Group([]g.Node{
							h.P(g.Text("You own the files and content you upload. You grant Arkive a limited license to host, store, and deliver your content solely to operate the service and provide the features you request.")),
							h.P(g.Text("You are responsible for the content you upload and share. Do not upload or share content that violates the Acceptable Use Policy.")),
							h.P(
								g.Text("If you believe content infringes copyright or violates our policies, submit a report at "),
								h.A(h.Href("/abuse"), h.Class("text-link"), g.Text("Copyright & Abuse Reporting")),
								g.Text("."),
							),
						}),
					),
					termsSection(
						"4. Data retention and archival",
						g.Group([]g.Node{
							h.P(g.Text("We keep account and file data for as long as your account is active. Accounts with no login or file activity for 30 days may be archived, and archived files are deleted after 7 days unless they are restored or upgraded.")),
							h.P(g.Text("We may notify you by email before archiving or deleting files.")),
						}),
					),
					termsSection(
						"5. Plans, payments, and rewards",
						g.Group([]g.Node{
							h.P(g.Text("We may display plan information on the website, but purchases are completed in the Android app through Google Play.")),
							h.P(g.Text("Rewarded ads may grant temporary storage boosts or rate limit increases. Rewards are promotional, may change, and may be revoked if abuse is detected.")),
						}),
					),
					termsSection(
						"6. Service availability",
						g.Group([]g.Node{
							h.P(g.Text("We work to keep Arkive available, but we do not guarantee uninterrupted service. We may modify, suspend, or discontinue features at any time.")),
						}),
					),
					termsSection(
						"7. Suspension and termination",
						g.Group([]g.Node{
							h.P(g.Text("We may suspend or terminate your access if you violate these Terms or the Acceptable Use Policy, or if we are required to do so by law.")),
						}),
					),
					termsSection(
						"8. Third-party services",
						g.Group([]g.Node{
							h.P(g.Text("Arkive integrates with third-party services such as Google Sign-In, Google Play billing, analytics, and ad providers. Your use of those services is subject to their terms and policies.")),
						}),
					),
					termsSection(
						"9. Disclaimers",
						g.Group([]g.Node{
							h.P(g.Text("Arkive is provided on an \"as is\" and \"as available\" basis. To the maximum extent permitted by law, we disclaim all warranties, express or implied.")),
						}),
					),
					termsSection(
						"10. Limitation of liability",
						g.Group([]g.Node{
							h.P(g.Text("To the maximum extent permitted by law, Arkive and its affiliates are not liable for indirect, incidental, special, consequential, or punitive damages, or for loss of data, profits, or revenue.")),
						}),
					),
					termsSection(
						"11. Indemnification",
						g.Group([]g.Node{
							h.P(g.Text("You agree to indemnify and hold Arkive and its affiliates harmless from claims arising out of your use of the service or violation of these Terms.")),
						}),
					),
					termsSection(
						"11. Governing law and disputes",
						g.Group([]g.Node{
							h.P(g.Text("These Terms are governed by the laws of the jurisdiction where Arkive is established. Disputes will be resolved in the courts of that jurisdiction.")),
						}),
					),
					termsSection(
						"12. Changes to these Terms",
						g.Group([]g.Node{
							h.P(g.Text("We may update these Terms from time to time. We will update the date above and, if changes are material, provide notice by posting on the service.")),
						}),
					),
					termsSection(
						"13. Contact",
						g.Group([]g.Node{
							h.P(
								g.Text("Questions? Email "),
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

func termsSection(title string, content g.Node) g.Node {
	return h.Section(
		h.Class("terms-section"),
		h.H2(g.Text(title)),
		content,
	)
}
