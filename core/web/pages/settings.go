package pages

import (
	"math"
	"strconv"
	"strings"
	"time"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/format"
	"arkive/pkg/validation"
)

type SettingsPageProps struct {
	Ctx             PageContext
	StorageSettings models.StorageSettings
	StorageGB       string
	Errors          validation.Errors
	Message         string
}

func SettingsPage(props SettingsPageProps) web.Page {
	user := props.Ctx.User
	brandName := ""
	email := ""
	instanceLabel := "Core"
	memberSince := "Unavailable"
	lastLogin := "Unavailable"
	usedStorage := "0 B"
	quotaStorage := "Unlimited"
	usagePercent := 0
	storageSettings := props.StorageSettings
	if storageSettings.Provider == "" {
		storageSettings.Provider = "local"
	}
	storageProviderLabel := strings.ToUpper(storageSettings.Provider)
	storageGB := props.StorageGB
	if storageGB == "" {
		storageGB = settingsStorageGB(storageSettings.MaxStorageBytes)
	}
	if user != nil {
		brandName = strings.TrimSpace(user.BrandName)
		email = strings.TrimSpace(user.Email)
		if !user.CreatedAt.IsZero() {
			memberSince = user.CreatedAt.Format("Jan 2, 2006")
		}
		if user.LastLoginAt != nil {
			lastLogin = user.LastLoginAt.Format(time.RFC1123)
		}
		usedStorage = format.Bytes(user.UsedBytes)
		if user.QuotaBytes > 0 && user.QuotaBytes != math.MaxInt64 {
			quotaStorage = format.Bytes(user.QuotaBytes)
			if user.UsedBytes > 0 {
				usagePercent = int((float64(user.UsedBytes) / float64(user.QuotaBytes)) * 100)
				if usagePercent > 100 {
					usagePercent = 100
				}
			}
		}
	}

	return web.Page{
		Title:      "Arkive · Settings",
		Robots:     RobotsNoIndex,
		CSS:        []string{"/web/pages/settings.css"},
		AuthLayout: true,
		User:       user,
		ActiveNav:  "settings",
		Body: h.Main(
			h.Class("settings-page"),
			h.Div(
				h.Class("container"),
				h.Div(
					h.Class("page-header"),
					h.Div(
						h.Class("page-title"),
						h.H1(g.Text("Settings")),
						h.P(g.Text("Manage storage for this self-hosted Core instance.")),
					),
				),
				h.Div(
					h.Class("settings-grid"),
					h.Div(
						h.Class("settings-column"),
						h.Div(
							g.Attr("id", "settings-account"),
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
						),
						h.Div(
							g.Attr("id", "settings-usage"),
							components.Card(components.CardProps{
								Title:    "Storage usage",
								Subtitle: "Usage and limits for this Core instance.",
								Class:    "settings-card",
								Body: []g.Node{
									h.Div(
										h.Class("settings-meta"),
										h.Div(
											h.Class("settings-meta-row"),
											h.Span(g.Text("Instance")),
											h.Span(h.Class("settings-badge"), g.Text(instanceLabel)),
										),
										h.Div(
											h.Class("settings-meta-row"),
											h.Span(g.Text("Storage used")),
											h.Span(g.Text(usedStorage+" / "+quotaStorage)),
										),
										h.Div(
											h.Class("settings-meta-row"),
											h.Span(g.Text("Provider")),
											h.Span(g.Text(storageProviderLabel)),
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
										g.Text("Storage limits are controlled by this Core instance configuration."),
									),
								},
							}),
						),
						h.Div(
							g.Attr("id", "settings-provider"),
							components.Card(components.CardProps{
								Title:    "Storage provider",
								Subtitle: "Change where new uploads are stored.",
								Class:    "settings-card",
								Body: []g.Node{
									storageSettingsForm(storageSettings, storageGB, props.Errors),
								},
							}),
						),
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
