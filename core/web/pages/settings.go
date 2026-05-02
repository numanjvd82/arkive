package pages

import (
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
		if user.QuotaBytes > 0 && user.QuotaBytes != 9223372036854775807 {
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
						components.Card(components.CardProps{
							Title:    "Storage provider",
							Subtitle: "Change where new uploads are stored.",
							Class:    "settings-card",
							Body: []g.Node{
								storageSettingsForm(storageSettings, storageGB, props.Errors),
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

func storageSettingsForm(settings models.StorageSettings, storageGB string, errors validation.Errors) g.Node {
	return h.Form(
		h.Class("settings-form"),
		g.Attr("method", "POST"),
		g.Attr("action", "/settings/storage"),
		g.If(
			validation.FieldError(errors, validation.GeneralKey) != "",
			h.P(h.Class("form-error"), g.Text(validation.FieldError(errors, validation.GeneralKey))),
		),
		h.Div(
			h.Class("storage-options"),
			storageRadio("local", "Local disk", settings.Provider == "local"),
			storageRadio("s3", "S3-compatible", settings.Provider == "s3"),
		),
		components.InputField(components.InputProps{
			Label:       "Storage limit in GB",
			Name:        "storage_gb",
			Type:        components.InputTypeNumber,
			Value:       storageGB,
			Description: "Use 0 for unlimited.",
			HelperText:  validation.FieldError(errors, "storage_gb"),
			HasError:    validation.FieldError(errors, "storage_gb") != "",
		}),
		h.Div(
			h.Class("settings-provider-fields local-panel"),
			components.InputField(components.InputProps{
				Label:      "Local path",
				Name:       "local_path",
				Type:       components.InputTypeText,
				Value:      settings.LocalPath,
				HelperText: validation.FieldError(errors, "local_path"),
				HasError:   validation.FieldError(errors, "local_path") != "",
			}),
		),
		h.Div(
			h.Class("settings-provider-fields s3-panel"),
			components.InputField(components.InputProps{
				Label:      "S3 access key",
				Name:       "s3_access_key_id",
				Type:       components.InputTypeText,
				Value:      settings.S3AccessKeyID,
				HelperText: validation.FieldError(errors, "s3_access_key_id"),
				HasError:   validation.FieldError(errors, "s3_access_key_id") != "",
			}),
			components.InputField(components.InputProps{
				Label:       "S3 secret key",
				Name:        "s3_secret_access_key",
				Type:        components.InputTypePassword,
				Placeholder: "Leave blank to keep current secret",
				HelperText:  validation.FieldError(errors, "s3_secret_access_key"),
				HasError:    validation.FieldError(errors, "s3_secret_access_key") != "",
			}),
			components.InputField(components.InputProps{
				Label: "S3 session token",
				Name:  "s3_session_token",
				Type:  components.InputTypePassword,
				Value: settings.S3SessionToken,
			}),
			components.InputField(components.InputProps{
				Label:      "S3 bucket",
				Name:       "s3_bucket",
				Type:       components.InputTypeText,
				Value:      settings.S3Bucket,
				HelperText: validation.FieldError(errors, "s3_bucket"),
				HasError:   validation.FieldError(errors, "s3_bucket") != "",
			}),
			components.InputField(components.InputProps{
				Label:      "S3 endpoint",
				Name:       "s3_endpoint",
				Type:       components.InputTypeText,
				Value:      settings.S3Endpoint,
				HelperText: validation.FieldError(errors, "s3_endpoint"),
				HasError:   validation.FieldError(errors, "s3_endpoint") != "",
			}),
			components.InputField(components.InputProps{
				Label: "S3 region",
				Name:  "s3_region",
				Type:  components.InputTypeText,
				Value: settings.S3Region,
			}),
		),
		components.Button(components.ButtonProps{
			Text:    "Save storage settings",
			Type:    "submit",
			Variant: "primary",
			Class:   "auth-submit",
		}),
	)
}

func storageRadio(value, label string, checked bool) g.Node {
	return h.Label(
		h.Class("storage-option compact"),
		h.Input(
			g.Attr("type", "radio"),
			g.Attr("name", "storage_provider"),
			g.Attr("value", value),
			g.If(checked, g.Attr("checked", "checked")),
		),
		h.Span(g.Text(label)),
	)
}

func settingsStorageGB(bytes int64) string {
	if bytes <= 0 || bytes == 9223372036854775807 {
		return "0"
	}
	return strconv.FormatInt(bytes/(1024*1024*1024), 10)
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
