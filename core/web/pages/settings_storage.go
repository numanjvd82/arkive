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
			Text:     "Save storage settings",
			Type:     "submit",
			Variant:  "primary",
			Class:    "auth-submit",
			BusyText: "Saving...",
		}),
	)
}

func uploadSettingsForm(settings models.UploadSettings, errors validation.Errors) g.Node {
	return h.Form(
		h.Class("settings-form"),
		g.Attr("method", "POST"),
		g.Attr("action", "/settings/uploads"),
		components.InputField(components.InputProps{Label: "Max queue items", Name: "max_queue_items", Type: components.InputTypeNumber, Value: strconv.Itoa(settings.MaxQueueItems), HelperText: validation.FieldError(errors, "max_queue_items"), HasError: validation.FieldError(errors, "max_queue_items") != ""}),
		components.InputField(components.InputProps{Label: "Part concurrency", Name: "part_concurrency", Type: components.InputTypeNumber, Value: strconv.Itoa(settings.PartConcurrency), Description: "Parallel upload parts per file. Mobile browsers are still capped lower.", HelperText: validation.FieldError(errors, "part_concurrency"), HasError: validation.FieldError(errors, "part_concurrency") != ""}),
		components.InputField(components.InputProps{Label: "Stale upload hours", Name: "stale_upload_hours", Type: components.InputTypeNumber, Value: strconv.Itoa(settings.StaleUploadHours), Description: "How long incomplete uploads and presigned upload URLs stay valid.", HelperText: validation.FieldError(errors, "stale_upload_hours"), HasError: validation.FieldError(errors, "stale_upload_hours") != ""}),
		components.Button(components.ButtonProps{Text: "Save upload settings", Type: "submit", Variant: "primary", Class: "auth-submit", BusyText: "Saving..."}),
	)
}

func previewSettingsForm(settings models.PreviewSettings, errors validation.Errors) g.Node {
	return h.Form(
		h.Class("settings-form"),
		g.Attr("method", "POST"),
		g.Attr("action", "/settings/previews"),
		components.InputField(components.InputProps{Label: "Image preview max MB", Name: "image_max_mb", Type: components.InputTypeNumber, Value: strconv.FormatInt(settings.ImageMaxBytes/(1024*1024), 10), HelperText: validation.FieldError(errors, "image_max_mb"), HasError: validation.FieldError(errors, "image_max_mb") != ""}),
		components.InputField(components.InputProps{Label: "Video preview max MB", Name: "video_max_mb", Type: components.InputTypeNumber, Value: strconv.FormatInt(settings.VideoMaxBytes/(1024*1024), 10), Description: "Used only for non-streaming fallback preview. Streaming playback is not capped by this value.", HelperText: validation.FieldError(errors, "video_max_mb"), HasError: validation.FieldError(errors, "video_max_mb") != ""}),
		components.InputField(components.InputProps{Label: "Text preview max MB", Name: "text_max_mb", Type: components.InputTypeNumber, Value: strconv.FormatInt(settings.TextMaxBytes/(1024*1024), 10), HelperText: validation.FieldError(errors, "text_max_mb"), HasError: validation.FieldError(errors, "text_max_mb") != ""}),
		components.Button(components.ButtonProps{Text: "Save preview settings", Type: "submit", Variant: "primary", Class: "auth-submit", BusyText: "Saving..."}),
	)
}

func settingsStorageGB(bytes int64) string {
	if bytes <= 0 || bytes == math.MaxInt64 {
		return "0"
	}
	return strconv.FormatInt(bytes/(1024*1024*1024), 10)
}
