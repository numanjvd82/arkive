package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/web"
	"arkive/core/web/components"
	"arkive/pkg/validation"
)

type SetupPageProps struct {
	Ctx               PageContext
	Errors            validation.Errors
	BrandName         string
	Email             string
	StorageProvider   string
	LocalPath         string
	StorageGB         string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3SessionToken    string
	S3Bucket          string
	S3Endpoint        string
	S3Region          string
	S3UsePathStyle    bool
}

func SetupPage(props SetupPageProps) web.Page {
	if props.StorageProvider == "" {
		props.StorageProvider = "local"
	}
	if props.LocalPath == "" {
		props.LocalPath = "./data/storage"
	}
	if props.StorageGB == "" {
		props.StorageGB = "0"
	}
	if props.S3Region == "" {
		props.S3Region = "auto"
	}
	return web.Page{
		Title:   "Arkive · Setup",
		Robots:  RobotsNoIndex,
		CSS:     []string{"/web/pages/setup.css"},
		Body:    setupBody(props),
		HideNav: true,
	}
}

func setupBody(props SetupPageProps) g.Node {
	generalError := validation.FieldError(props.Errors, validation.GeneralKey)

	return h.Main(
		h.Class("setup-page"),
		h.Div(
			h.Class("setup-shell"),
			h.Section(
				h.Class("setup-intro"),
				h.P(h.Class("label"), g.Text("Arkive Core")),
				h.H1(g.Text("Set up your instance")),
				h.P(g.Text("Create the admin account and choose where encrypted files are stored.")),
			),
			h.Form(
				h.Class("setup-form"),
				g.Attr("method", "POST"),
				g.If(
					generalError != "",
					h.P(h.Class("form-error"), g.Text(generalError)),
				),
				stepHeader("1", "Admin account"),
				h.Div(
					h.Class("setup-panel"),
					components.InputField(components.InputProps{
						Label:       "Instance name",
						Name:        "brand_name",
						Type:        components.InputTypeText,
						Placeholder: "Arkive",
						Value:       props.BrandName,
						Required:    true,
						HelperText:  validation.FieldError(props.Errors, "brand_name"),
						HasError:    validation.FieldError(props.Errors, "brand_name") != "",
					}),
					components.InputField(components.InputProps{
						Label:       "Admin email",
						Name:        "email",
						Type:        components.InputTypeEmail,
						Placeholder: "you@example.com",
						Value:       props.Email,
						Required:    true,
						HelperText:  validation.FieldError(props.Errors, "email"),
						HasError:    validation.FieldError(props.Errors, "email") != "",
					}),
					components.InputField(components.InputProps{
						Label:       "Password",
						Name:        "password",
						Type:        components.InputTypePassword,
						Placeholder: "Create a password",
						Required:    true,
						HelperText:  validation.FieldError(props.Errors, "password"),
						HasError:    validation.FieldError(props.Errors, "password") != "",
					}),
					components.InputField(components.InputProps{
						Label:       "Confirm password",
						Name:        "confirm_password",
						Type:        components.InputTypePassword,
						Placeholder: "Confirm your password",
						Required:    true,
						HelperText:  validation.FieldError(props.Errors, "confirm_password"),
						HasError:    validation.FieldError(props.Errors, "confirm_password") != "",
					}),
				),
				stepHeader("2", "Storage"),
				h.Div(
					h.Class("setup-panel"),
					h.Div(
						h.Class("storage-options"),
						storageOption("local", "Local disk", "Use a folder on this server.", props.StorageProvider == "local"),
						storageOption("s3", "S3-compatible", "Use R2, Backblaze, Wasabi, AWS, or compatible storage.", props.StorageProvider == "s3"),
					),
					components.InputField(components.InputProps{
						Label:       "Storage limit in GB",
						Name:        "storage_gb",
						Type:        components.InputTypeNumber,
						Placeholder: "0",
						Value:       props.StorageGB,
						Required:    true,
						Description: "Use 0 for unlimited.",
						HelperText:  validation.FieldError(props.Errors, "storage_gb"),
						HasError:    validation.FieldError(props.Errors, "storage_gb") != "",
					}),
					h.Div(
						h.Class("provider-grid"),
						h.FieldSet(
							h.Class("provider-panel local-panel"),
							h.Legend(g.Text("Local disk")),
							components.InputField(components.InputProps{
								Label:       "Storage path",
								Name:        "local_path",
								Type:        components.InputTypeText,
								Placeholder: "/var/lib/arkive/files",
								Value:       props.LocalPath,
								HelperText:  validation.FieldError(props.Errors, "local_path"),
								HasError:    validation.FieldError(props.Errors, "local_path") != "",
							}),
						),
						h.FieldSet(
							h.Class("provider-panel s3-panel"),
							h.Legend(g.Text("S3-compatible")),
							components.InputField(components.InputProps{
								Label:      "Access key",
								Name:       "s3_access_key_id",
								Type:       components.InputTypeText,
								Value:      props.S3AccessKeyID,
								HelperText: validation.FieldError(props.Errors, "s3_access_key_id"),
								HasError:   validation.FieldError(props.Errors, "s3_access_key_id") != "",
							}),
							components.InputField(components.InputProps{
								Label:      "Secret key",
								Name:       "s3_secret_access_key",
								Type:       components.InputTypePassword,
								Value:      props.S3SecretAccessKey,
								HelperText: validation.FieldError(props.Errors, "s3_secret_access_key"),
								HasError:   validation.FieldError(props.Errors, "s3_secret_access_key") != "",
							}),
							components.InputField(components.InputProps{
								Label: "Session token",
								Name:  "s3_session_token",
								Type:  components.InputTypePassword,
								Value: props.S3SessionToken,
							}),
							components.InputField(components.InputProps{
								Label:      "Bucket",
								Name:       "s3_bucket",
								Type:       components.InputTypeText,
								Value:      props.S3Bucket,
								HelperText: validation.FieldError(props.Errors, "s3_bucket"),
								HasError:   validation.FieldError(props.Errors, "s3_bucket") != "",
							}),
							components.InputField(components.InputProps{
								Label:       "Endpoint",
								Name:        "s3_endpoint",
								Type:        components.InputTypeText,
								Placeholder: "https://account.r2.cloudflarestorage.com",
								Value:       props.S3Endpoint,
								HelperText:  validation.FieldError(props.Errors, "s3_endpoint"),
								HasError:    validation.FieldError(props.Errors, "s3_endpoint") != "",
							}),
							components.InputField(components.InputProps{
								Label:       "Region",
								Name:        "s3_region",
								Type:        components.InputTypeText,
								Placeholder: "auto",
								Value:       props.S3Region,
							}),
						),
					),
				),
				stepHeader("3", "Finish"),
				h.Div(
					h.Class("setup-actions"),
					components.Button(components.ButtonProps{
						Text:    "Create instance",
						Type:    "submit",
						Variant: "primary",
						Class:   "auth-submit",
					}),
				),
			),
		),
	)
}

func stepHeader(number, title string) g.Node {
	return h.Div(
		h.Class("setup-step"),
		h.Span(h.Class("setup-step-number"), g.Text(number)),
		h.H2(g.Text(title)),
	)
}

func storageOption(value, title, description string, checked bool) g.Node {
	return h.Label(
		h.Class("storage-option"),
		h.Input(
			g.Attr("type", "radio"),
			g.Attr("name", "storage_provider"),
			g.Attr("value", value),
			g.If(checked, g.Attr("checked", "checked")),
		),
		h.Span(
			h.Class("storage-option-copy"),
			h.Strong(g.Text(title)),
			h.Span(g.Text(description)),
		),
	)
}
