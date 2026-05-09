package pages

import (
	"math"
	"strconv"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web/components"
	"arkive/pkg/validation"
)

func storageSettingsForm(settings models.StorageSettings, storageGB string, errors validation.Errors) g.Node {
	return h.Form(
		h.Class("settings-form"),
		g.Attr("method", "POST"),
		g.Attr("action", "/settings/storage"),
		g.If(
			validation.FieldError(errors, validation.GeneralKey) != "",
			h.P(h.Class("form-error"), g.Text(validation.FieldError(errors, validation.GeneralKey))),
		),
		components.StorageSelector("storage_provider",
			components.StorageSelectorOption{Value: "local", Label: "Local", Checked: settings.Provider == "local"},
			components.StorageSelectorOption{Value: "s3", Label: "S3 Compatible", Checked: settings.Provider == "s3"},
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

func emailSettingsForm(settings models.EmailSettings, errors validation.Errors) g.Node {
	return h.Form(
		h.Class("settings-form"),
		g.Attr("method", "POST"),
		g.Attr("action", "/settings/email"),
		components.InputField(components.InputProps{Label: "Provider", Name: "email_provider", Type: components.InputTypeText, Value: settings.Provider}),
		components.InputField(components.InputProps{Label: "From address", Name: "email_from", Type: components.InputTypeText, Value: settings.From, HelperText: validation.FieldError(errors, "email_from"), HasError: validation.FieldError(errors, "email_from") != ""}),
		components.InputField(components.InputProps{Label: "Public base URL", Name: "public_base_url", Type: components.InputTypeText, Value: settings.PublicBaseURL}),
		components.InputField(components.InputProps{Label: "SMTP host", Name: "smtp_host", Type: components.InputTypeText, Value: settings.SMTPHost, HelperText: validation.FieldError(errors, "smtp_host"), HasError: validation.FieldError(errors, "smtp_host") != ""}),
		components.InputField(components.InputProps{Label: "SMTP port", Name: "smtp_port", Type: components.InputTypeNumber, Value: strconv.Itoa(settings.SMTPPort), HelperText: validation.FieldError(errors, "smtp_port"), HasError: validation.FieldError(errors, "smtp_port") != ""}),
		components.InputField(components.InputProps{Label: "SMTP user", Name: "smtp_user", Type: components.InputTypeText, Value: settings.SMTPUser}),
		components.InputField(components.InputProps{Label: "SMTP pass", Name: "smtp_pass", Type: components.InputTypePassword, Value: settings.SMTPPass}),
		components.Button(components.ButtonProps{Text: "Save email settings", Type: "submit", Variant: "primary", Class: "auth-submit"}),
	)
}

func uploadSettingsForm(settings models.UploadSettings, errors validation.Errors) g.Node {
	return h.Form(
		h.Class("settings-form"),
		g.Attr("method", "POST"),
		g.Attr("action", "/settings/uploads"),
		components.InputField(components.InputProps{Label: "Max queue items", Name: "max_queue_items", Type: components.InputTypeNumber, Value: strconv.Itoa(settings.MaxQueueItems), HelperText: validation.FieldError(errors, "max_queue_items"), HasError: validation.FieldError(errors, "max_queue_items") != ""}),
		components.Button(components.ButtonProps{Text: "Save upload settings", Type: "submit", Variant: "primary", Class: "auth-submit"}),
	)
}

func settingsStorageGB(bytes int64) string {
	if bytes <= 0 || bytes == math.MaxInt64 {
		return "0"
	}
	return strconv.FormatInt(bytes/(1024*1024*1024), 10)
}
