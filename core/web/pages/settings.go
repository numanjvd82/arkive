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
	EmailSettings   models.EmailSettings
	UploadSettings  models.UploadSettings
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
	emailSettings := props.EmailSettings
	uploadSettings := props.UploadSettings
	if storageSettings.Provider == "" {
		storageSettings.Provider = "local"
	}
	if emailSettings.Provider == "" {
		emailSettings.Provider = "noop"
	}
	if uploadSettings.MaxQueueItems == 0 {
		uploadSettings.MaxQueueItems = 300
	}
	storageProviderLabel := strings.ToUpper(storageSettings.Provider)
	emailProviderLabel := strings.ToUpper(emailSettings.Provider)
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
		Title:              "Arkive · Settings",
		Robots:             RobotsNoIndex,
		CSS:                []string{"/web/pages/settings.css"},
		AuthLayout:         true,
		RequireVaultUnlock: true,
		User:               user,
		ActiveNav:          "settings",
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
					h.Aside(
						h.Class("settings-tabs"),
						settingsTabLink("settings-account", "Account"),
						settingsTabLink("settings-provider", "Storage Provider"),
						settingsTabLink("settings-email", "Email"),
						settingsTabLink("settings-upload", "Uploads"),
						settingsTabLink("settings-security", "Security"),
					),
					h.Div(
						h.Class("settings-content"),
						h.Section(
							h.Class("settings-panel settings-panel-default"),
							g.Attr("id", "settings-account"),
							h.Div(
								h.Class("settings-panel-header"),
								h.Div(
									h.Class("settings-panel-title"),
									h.H2(g.Text("Account Overview")),
									h.P(g.Text("Current account, workspace, and storage details for this Core instance.")),
								),
							),
							h.Div(
								h.Class("settings-stack"),
								components.Card(components.CardProps{
									Title:    "Account details",
									Subtitle: "Identity and workspace metadata for the active admin session.",
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
									Title:    "Storage quota",
									Subtitle: "Usage and limits for this single-user Core deployment.",
									Class:    "settings-card",
									Body: []g.Node{
										h.Div(
											h.Class("settings-quota-head"),
											h.Div(
												h.Class("settings-quota-copy"),
												h.Span(h.Class("settings-overline"), g.Text("Current usage")),
												h.Strong(g.Text(usedStorage)),
											),
											h.Div(
												h.Class("settings-quota-total"),
												h.Strong(g.Text(quotaStorage)),
												h.Span(g.Text("configured limit")),
											),
										),
										h.Div(
											h.Class("settings-progress"),
											h.Div(
												h.Class("settings-progress-track"),
												h.Span(
													h.Class("settings-progress-bar"),
													g.Attr("style", "width: "+formatPercent(usagePercent)),
												),
											),
										),
										h.Div(
											h.Class("settings-info-grid"),
											settingsInfoTile("Instance", instanceLabel, "Self-hosted mode"),
											settingsInfoTile("Provider", storageProviderLabel, "Active upload backend"),
											settingsInfoTile("Utilization", formatPercent(usagePercent), "Based on the configured quota"),
										),
										h.P(
											h.Class("settings-note"),
											g.Text("Storage limits are controlled by instance configuration. Set the limit to 0 to allow unlimited storage."),
										),
									},
								}),
							),
						),
						h.Section(
							h.Class("settings-panel"),
							g.Attr("id", "settings-provider"),
							h.Div(
								h.Class("settings-panel-header"),
								h.Div(
									h.Class("settings-panel-title"),
									h.H2(g.Text("Storage Provider")),
									h.P(g.Text("Choose between local disk and S3-compatible storage for new uploads.")),
								),
							),
							h.Div(
								h.Class("settings-stack"),
								components.Card(components.CardProps{
									Title:    "Provider status",
									Subtitle: "The active backend determines where encrypted blobs are stored.",
									Class:    "settings-card",
									Body: []g.Node{
										storageProviderSummary(storageSettings),
									},
								}),
								components.Card(components.CardProps{
									Title:    "Configuration",
									Subtitle: "Update provider settings for this Core instance.",
									Class:    "settings-card",
									Body: []g.Node{
										storageSettingsForm(storageSettings, storageGB, props.Errors),
									},
								}),
							),
						),
						h.Section(
							h.Class("settings-panel"),
							g.Attr("id", "settings-email"),
							h.Div(
								h.Class("settings-panel-header"),
								h.Div(
									h.Class("settings-panel-title"),
									h.H2(g.Text("Email")),
									h.P(g.Text("Configure the instance mailer from settings instead of environment variables.")),
								),
							),
							h.Div(
								h.Class("settings-stack"),
								components.Card(components.CardProps{
									Title:    "Mailer status",
									Subtitle: "Active email provider for auth and recovery flows.",
									Class:    "settings-card",
									Body: []g.Node{
										h.Div(
											h.Class("settings-provider-status"),
											h.Span(h.Class("settings-badge"), g.Text(emailProviderLabel)),
											h.H4(g.Text("Mailer configured from settings")),
											h.P(g.Text("The app reads provider, sender, host, and credentials from instance settings.")),
										),
									},
								}),
								components.Card(components.CardProps{
									Title:    "Configuration",
									Subtitle: "Update email delivery settings for this Core instance.",
									Class:    "settings-card",
									Body: []g.Node{
										emailSettingsForm(emailSettings, props.Errors),
									},
								}),
							),
						),
						h.Section(
							h.Class("settings-panel"),
							g.Attr("id", "settings-upload"),
							h.Div(
								h.Class("settings-panel-header"),
								h.Div(
									h.Class("settings-panel-title"),
									h.H2(g.Text("Uploads")),
									h.P(g.Text("Configure upload queue limits from settings instead of environment variables.")),
								),
							),
							h.Div(
								h.Class("settings-stack"),
								components.Card(components.CardProps{
									Title:    "Current limits",
									Subtitle: "Upload limits applied by the server.",
									Class:    "settings-card",
									Body: []g.Node{
										h.Div(
											h.Class("settings-meta"),
											h.Div(h.Class("settings-meta-row"), h.Span(g.Text("Queue items")), h.Span(g.Text(strconv.Itoa(uploadSettings.MaxQueueItems)))),
										),
									},
								}),
								components.Card(components.CardProps{
									Title:    "Configuration",
									Subtitle: "Update upload queue settings for this Core instance.",
									Class:    "settings-card",
									Body: []g.Node{
										uploadSettingsForm(uploadSettings, props.Errors),
									},
								}),
							),
						),
						h.Section(
							h.Class("settings-panel"),
							g.Attr("id", "settings-security"),
							h.Div(
								h.Class("settings-panel-header"),
								h.Div(
									h.Class("settings-panel-title"),
									h.H2(g.Text("Security")),
									h.P(g.Text("Reserved for future authentication and instance hardening controls.")),
								),
							),
							components.Card(components.CardProps{
								Title:    "Coming soon",
								Subtitle: "This section will hold security-focused settings as Core expands.",
								Class:    "settings-card",
								Body: []g.Node{
									h.Ul(
										h.Class("settings-list"),
										h.Li(g.Text("Password and session controls")),
										h.Li(g.Text("Email verification and recovery options")),
										h.Li(g.Text("Instance access and audit details")),
									),
								},
							}),
						),
					),
				),
			),
		),
	}
}

func settingsTabLink(target, label string) g.Node {
	return h.A(
		h.Class("settings-tab"),
		h.Href("#"+target),
		g.Text(label),
	)
}

func settingsInfoTile(label, value, hint string) g.Node {
	return h.Div(
		h.Class("settings-info-tile"),
		h.Span(h.Class("settings-info-label"), g.Text(label)),
		h.Strong(g.Text(value)),
		h.P(g.Text(hint)),
	)
}

func storageProviderSummary(settings models.StorageSettings) g.Node {
	provider := strings.ToLower(strings.TrimSpace(settings.Provider))
	if provider == "" {
		provider = "local"
	}

	if provider == "s3" {
		return h.Div(
			h.Class("settings-provider-status"),
			h.Span(h.Class("settings-badge"), g.Text("S3")),
			h.H4(g.Text("S3-compatible storage is active")),
			h.P(g.Text("Uploads are stored through the configured S3-compatible provider without exposing provider-specific identifiers here.")),
		)
	}

	return h.Div(
		h.Class("settings-provider-status"),
		h.Span(h.Class("settings-badge"), g.Text("LOCAL")),
		h.H4(g.Text("Local storage is active")),
		h.P(g.Text("Uploads are stored on the instance filesystem without exposing path-level details in this overview.")),
	)
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
