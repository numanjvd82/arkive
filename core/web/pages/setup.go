package pages

import (
	lucide "github.com/eduardolat/gomponents-lucide"
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
		JS:      []string{"/static/setup.js"},
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
			h.Form(
				h.Class("setup-form"),
				g.Attr("method", "POST"),
				g.Attr("data-setup-vault-form", "true"),
				h.Input(h.Type("hidden"), h.Name("vault_salt"), g.Attr("data-vault-salt-input", "true")),
				h.Input(h.Type("hidden"), h.Name("encrypted_master_key"), g.Attr("data-encrypted-master-key-input", "true")),
				h.Div(
					h.Class("setup-header"),
					h.Span(
						h.Class("setup-header-icon"),
						lucide.Server(
							h.Class("setup-lucide setup-lucide-header"),
							g.Attr("aria-hidden", "true"),
						),
					),
					h.Div(
						h.Class("setup-header-copy"),
						h.H1(g.Text("Initialize Arkive Core")),
						h.P(g.Text("Configure your self-hosted encrypted file server.")),
					),
				),
				h.Div(
					h.Class("setup-body"),
					g.If(
						generalError != "",
						h.P(h.Class("form-error"), g.Text(generalError)),
					),
					setupSection(
						"1",
						"Instance Details",
						lucide.BadgeInfo(
							h.Class("setup-lucide"),
							g.Attr("aria-hidden", "true"),
						),
						h.Div(
							h.Class("setup-grid setup-grid-two"),
							components.InputField(components.InputProps{
								Label:       "Instance name",
								Name:        "brand_name",
								Type:        components.InputTypeText,
								Placeholder: "Sovereign-01",
								Value:       props.BrandName,
								Required:    true,
								HelperText:  validation.FieldError(props.Errors, "brand_name"),
								HasError:    validation.FieldError(props.Errors, "brand_name") != "",
								InputClass:  "setup-input mono",
							}),
							components.InputField(components.InputProps{
								Label:       "Admin email",
								Name:        "email",
								Type:        components.InputTypeEmail,
								Placeholder: "admin@domain.local",
								Value:       props.Email,
								Required:    true,
								HelperText:  validation.FieldError(props.Errors, "email"),
								HasError:    validation.FieldError(props.Errors, "email") != "",
								InputClass:  "setup-input mono",
							}),
						),
					),
					setupSection(
						"2",
						"Security",
						lucide.Lock(
							h.Class("setup-lucide"),
							g.Attr("aria-hidden", "true"),
						),
						h.Div(
							h.Class("setup-warning"),
							h.Span(
								h.Class("setup-warning-icon"),
								lucide.ShieldAlert(
									h.Class("setup-lucide setup-lucide-warning"),
									g.Attr("aria-hidden", "true"),
								),
							),
							h.P(
								g.Text("Arkive Core uses zero-knowledge encryption. This password derives your master encryption key. "),
								h.Strong(g.Text("It cannot be recovered if lost.")),
							),
						),
						h.Div(
							h.Class("setup-grid setup-grid-two"),
							components.InputField(components.InputProps{
								Label:       "Master password",
								Name:        "password",
								Type:        components.InputTypePassword,
								Placeholder: "Create a password",
								Required:    true,
								HelperText:  validation.FieldError(props.Errors, "password"),
								HasError:    validation.FieldError(props.Errors, "password") != "",
								InputClass:  "setup-input mono",
							}),
							components.InputField(components.InputProps{
								Label:       "Confirm password",
								Name:        "confirm_password",
								Type:        components.InputTypePassword,
								Placeholder: "Confirm your password",
								Required:    true,
								HelperText:  validation.FieldError(props.Errors, "confirm_password"),
								HasError:    validation.FieldError(props.Errors, "confirm_password") != "",
								InputClass:  "setup-input mono",
							}),
						),
					),
					setupSection(
						"3",
						"Storage",
						lucide.Database(
							h.Class("setup-lucide"),
							g.Attr("aria-hidden", "true"),
						),
						h.Div(
							h.Class("storage-section"),
							components.StorageSelector("storage_provider",
								components.StorageSelectorOption{Value: "local", Label: "Local", Checked: props.StorageProvider == "local"},
								components.StorageSelectorOption{Value: "s3", Label: "S3 Compatible", Checked: props.StorageProvider == "s3"},
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
								InputClass:  "setup-input mono",
							}),
							h.Div(
								h.Class("provider-stack"),
								h.Div(
									h.Class("provider-panel local-fields"),
									components.InputField(components.InputProps{
										Label:       "Storage path",
										Name:        "local_path",
										Type:        components.InputTypeText,
										Placeholder: "/var/lib/arkive/files",
										Value:       props.LocalPath,
										HelperText:  validation.FieldError(props.Errors, "local_path"),
										HasError:    validation.FieldError(props.Errors, "local_path") != "",
										InputClass:  "setup-input mono",
									}),
								),
								h.Div(
									h.Class("provider-panel s3-fields"),
									h.Div(
										h.Class("setup-grid"),
										h.Div(
											h.Class("setup-grid-span"),
											components.InputField(components.InputProps{
												Label:       "Endpoint URL",
												Name:        "s3_endpoint",
												Type:        components.InputTypeText,
												Placeholder: "https://s3.region.amazonaws.com",
												Value:       props.S3Endpoint,
												HelperText:  validation.FieldError(props.Errors, "s3_endpoint"),
												HasError:    validation.FieldError(props.Errors, "s3_endpoint") != "",
												InputClass:  "setup-input mono",
											}),
										),
										components.InputField(components.InputProps{
											Label:       "Bucket name",
											Name:        "s3_bucket",
											Type:        components.InputTypeText,
											Placeholder: "arkive-vault-01",
											Value:       props.S3Bucket,
											HelperText:  validation.FieldError(props.Errors, "s3_bucket"),
											HasError:    validation.FieldError(props.Errors, "s3_bucket") != "",
											InputClass:  "setup-input mono",
										}),
										components.InputField(components.InputProps{
											Label:       "Region",
											Name:        "s3_region",
											Type:        components.InputTypeText,
											Placeholder: "us-east-1",
											Value:       props.S3Region,
											InputClass:  "setup-input mono",
										}),
										components.InputField(components.InputProps{
											Label:      "Access key",
											Name:       "s3_access_key_id",
											Type:       components.InputTypeText,
											Value:      props.S3AccessKeyID,
											HelperText: validation.FieldError(props.Errors, "s3_access_key_id"),
											HasError:   validation.FieldError(props.Errors, "s3_access_key_id") != "",
											InputClass: "setup-input mono",
										}),
										components.InputField(components.InputProps{
											Label:      "Secret key",
											Name:       "s3_secret_access_key",
											Type:       components.InputTypePassword,
											Value:      props.S3SecretAccessKey,
											HelperText: validation.FieldError(props.Errors, "s3_secret_access_key"),
											HasError:   validation.FieldError(props.Errors, "s3_secret_access_key") != "",
											InputClass: "setup-input mono",
										}),
										components.InputField(components.InputProps{
											Label:      "Session token",
											Name:       "s3_session_token",
											Type:       components.InputTypePassword,
											Value:      props.S3SessionToken,
											InputClass: "setup-input mono",
										}),
										h.Label(
											h.Class("setup-checkbox"),
											h.Input(
												g.Attr("type", "checkbox"),
												g.Attr("name", "s3_use_path_style"),
												g.If(props.S3UsePathStyle, g.Attr("checked", "checked")),
											),
											h.Span(g.Text("Use path-style addressing")),
										),
									),
								),
							),
						),
					),
				),
				h.Div(
					h.Class("setup-footer"),
					h.Button(
						h.Class("setup-cancel"),
						h.Type("reset"),
						g.Text("Cancel"),
					),
					h.Button(
						h.Class("setup-submit"),
						h.Type("submit"),
						g.Attr("data-busy-text", "Initializing..."),
						h.Span(g.Text("Initialize Arkive Core")),
						lucide.Rocket(
							h.Class("setup-lucide setup-lucide-submit"),
							g.Attr("aria-hidden", "true"),
						),
					),
				),
			),
		),
	)
}

func setupSection(number, title string, icon g.Node, children ...g.Node) g.Node {
	nodes := []g.Node{
		h.Div(
			h.Class("setup-section-header"),
			h.Span(h.Class("setup-section-icon"), icon),
			h.H2(g.Text(number+". "+title)),
		),
	}
	nodes = append(nodes, children...)
	return h.Section(
		h.Class("setup-section"),
		g.Group(nodes),
	)
}
