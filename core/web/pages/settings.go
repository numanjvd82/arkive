package pages

import (
	"strconv"
	"strings"
	"time"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
)

type SettingsPageProps struct {
	Ctx            PageContext
	FileCount      int64
	FileLimitLabel string
}

func SettingsPage(props SettingsPageProps) web.Page {
	user := props.Ctx.User
	brandName := ""
	email := ""
	planName := "Free"
	memberSince := "Unavailable"
	lastLogin := "Unavailable"
	usedStorage := "0 B"
	quotaStorage := "Unlimited"
	usagePercent := 0
	fileLimit := props.FileLimitLabel
	fileCountLabel := "0"
	retentionLabel := "While active (archive after inactivity)"
	if user != nil {
		brandName = strings.TrimSpace(user.BrandName)
		email = strings.TrimSpace(user.Email)
		if user.IsPremium {
			planName = "Premium"
			retentionLabel = "While active (archive after inactivity)"
			if fileLimit == "" {
				fileLimit = "Unlimited"
			}
		}
		if !user.CreatedAt.IsZero() {
			memberSince = user.CreatedAt.Format("Jan 2, 2006")
		}
		if user.LastLoginAt != nil {
			lastLogin = user.LastLoginAt.Format(time.RFC1123)
		}
		usedStorage = format.Bytes(user.UsedBytes)
		if user.QuotaBytes > 0 {
			quotaStorage = format.Bytes(user.QuotaBytes)
			if user.UsedBytes > 0 {
				usagePercent = int((float64(user.UsedBytes) / float64(user.QuotaBytes)) * 100)
				if usagePercent > 100 {
					usagePercent = 100
				}
			}
		}
	}
	if props.FileCount > 0 {
		fileCountLabel = format.Commas(props.FileCount)
	}
	if fileLimit == "" {
		fileLimit = "Unlimited"
	}

	return web.Page{
		Title:      "Arkive · Settings",
		Robots:     RobotsNoIndex,
		CSS:        []string{"/web/pages/settings.css"},
		AuthLayout: true,
		User:       user,
		Body: h.Main(
			h.Class("settings-page"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("page-header"),
					h.Div(
						h.Class("page-title"),
						h.H1(g.Text("Settings")),
						h.P(g.Text("Read-only account details for now. More settings are coming soon.")),
					),
				),
				h.Div(
					h.Class("settings-grid"),
					h.Div(
						h.Class("settings-column"),
						components.Card(components.CardProps{
							Title:    "Account details",
							Subtitle: "Your Arkive workspace profile.",
							Class:    "settings-card",
							Body: []g.Node{
								h.Div(
									h.Class("settings-meta"),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Workspace")),
										h.Span(g.Text(displayOrDash(brandName))),
									),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Email")),
										h.Span(g.Text(displayOrDash(email))),
									),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Member since")),
										h.Span(g.Text(displayOrDash(memberSince))),
									),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Last login")),
										h.Span(g.Text(displayOrDash(lastLogin))),
									),
								),
							},
						}),
						components.Card(components.CardProps{
							Title:    "Plan & storage",
							Subtitle: "Usage and limits pulled from your workspace.",
							Class:    "settings-card",
							Body: []g.Node{
								h.Div(
									h.Class("settings-meta"),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Plan")),
										h.Span(h.Class("settings-badge"), g.Text(planName)),
									),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Storage used")),
										h.Span(g.Text(usedStorage+" / "+quotaStorage)),
									),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("File limit")),
										h.Span(g.Text(fileLimit)),
									),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Files")),
										h.Span(g.Text(fileCountLabel+" / "+fileLimit)),
									),
									h.Div(
										h.Class("settings-meta-row"),
										h.Span(g.Text("Retention")),
										h.Span(g.Text(retentionLabel)),
									),
								),
								h.Div(
									h.Class("settings-progress"),
									h.Span(h.Class("settings-progress-label"), g.Text("Storage usage")),
									h.Div(
										h.Class("settings-progress-track"),
										h.Span(
											h.Class("settings-progress-bar"),
											g.Attr("style", "width: "+formatPercent(usagePercent)),
										),
									),
								),
								h.P(
									h.Class("settings-note"),
									g.Text("Free accounts can store up to 10 GB and 10,000 files. Accounts with no login or file activity for 30 days may be archived; archived files are deleted after 7 days unless restored or upgraded."),
								),
							},
						}),
					),
					h.Aside(
						h.Class("settings-column settings-side"),
						components.Card(components.CardProps{
							Title:    "Coming soon",
							Subtitle: "More settings are on the way.",
							Class:    "settings-card",
							Body: []g.Node{
								h.Ul(
									h.Class("settings-list"),
									h.Li(g.Text("Profile edits and password updates.")),
									h.Li(g.Text("Notification and sharing defaults.")),
									h.Li(g.Text("Two-factor authentication controls.")),
								),
							},
						}),
					),
				),
			),
		),
	}
}

func formatPercent(value int) string {
	if value <= 0 {
		return "0%"
	}
	if value >= 100 {
		return "100%"
	}
	return strconv.Itoa(value) + "%"
}

func displayOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "Unavailable"
	}
	return value
}
