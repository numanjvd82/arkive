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
	if bytes <= 0 || bytes == math.MaxInt64 {
		return "0"
	}
	return strconv.FormatInt(bytes/(1024*1024*1024), 10)
}
