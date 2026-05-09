package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	settingssvc "arkive/core/services/settings"
	"arkive/core/services/uploads"
	"arkive/core/web"
	pages "arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/errs"
	"arkive/pkg/validation"
)

type storageSettingsForm struct {
	StorageProvider   string `form:"storage_provider"`
	LocalPath         string `form:"local_path"`
	StorageGB         string `form:"storage_gb"`
	S3AccessKeyID     string `form:"s3_access_key_id"`
	S3SecretAccessKey string `form:"s3_secret_access_key"`
	S3SessionToken    string `form:"s3_session_token"`
	S3Bucket          string `form:"s3_bucket"`
	S3Endpoint        string `form:"s3_endpoint"`
	S3Region          string `form:"s3_region"`
	S3UsePathStyle    string `form:"s3_use_path_style"`
}

type emailSettingsForm struct {
	Provider      string `form:"email_provider"`
	From          string `form:"email_from"`
	PublicBaseURL string `form:"public_base_url"`
	SMTPHost      string `form:"smtp_host"`
	SMTPPort      string `form:"smtp_port"`
	SMTPUser      string `form:"smtp_user"`
	SMTPPass      string `form:"smtp_pass"`
}

type uploadSettingsForm struct {
	MaxQueueItems string `form:"max_queue_items"`
}

func WebSettings(uploadService *uploads.Service, settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err := uploadService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		storageSettings, _ := settingsService.StorageSettings(c.Request.Context())
		emailSettings, _ := settingsService.EmailSettings(c.Request.Context())
		uploadSettings, _ := settingsService.UploadSettings(c.Request.Context())
		web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
			Ctx:             pages.ContextWithUser(user),
			StorageSettings: storageSettings,
			EmailSettings:   emailSettings,
			UploadSettings:  uploadSettings,
		}))
	}
}

func WebSettingsStoragePost(settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var form storageSettingsForm
		if err := c.ShouldBind(&form); err != nil {
			renderSettingsStorage(c, user, models.StorageSettings{}, validation.Errors{validation.GeneralKey: "Please fill out the form."})
			return
		}

		settings, validationErrors, err := settingsService.UpdateStorageSettings(c.Request.Context(), user.ID, settingsInputFromForm(form))
		if validationErrors != nil && validationErrors.HasAny() {
			renderSettingsStorage(c, user, settings, validationErrors)
			return
		}
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Redirect(http.StatusSeeOther, "/settings?msg=storage-updated")
	}
}

func WebSettingsEmailPost(settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var form emailSettingsForm
		if err := c.ShouldBind(&form); err != nil {
			renderSettingsEmail(c, user, models.EmailSettings{}, validation.Errors{validation.GeneralKey: "Please fill out the form."})
			return
		}

		settings, validationErrors, err := settingsService.UpdateEmailSettings(c.Request.Context(), user.ID, emailSettingsFromForm(form))
		if validationErrors != nil && validationErrors.HasAny() {
			renderSettingsEmail(c, user, settings, validationErrors)
			return
		}
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Redirect(http.StatusSeeOther, "/settings?msg=email-updated")
	}
}

func WebSettingsUploadPost(settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var form uploadSettingsForm
		if err := c.ShouldBind(&form); err != nil {
			renderSettingsUpload(c, user, models.UploadSettings{}, validation.Errors{validation.GeneralKey: "Please fill out the form."})
			return
		}

		settings, validationErrors, err := settingsService.UpdateUploadSettings(c.Request.Context(), user.ID, uploadSettingsFromForm(form))
		if validationErrors != nil && validationErrors.HasAny() {
			renderSettingsUpload(c, user, settings, validationErrors)
			return
		}
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Redirect(http.StatusSeeOther, "/settings?msg=upload-updated")
	}
}

func settingsInputFromForm(form storageSettingsForm) settingssvc.StorageInput {
	return settingssvc.StorageInput{
		Provider:          form.StorageProvider,
		LocalPath:         form.LocalPath,
		StorageGB:         form.StorageGB,
		S3AccessKeyID:     form.S3AccessKeyID,
		S3SecretAccessKey: form.S3SecretAccessKey,
		S3SessionToken:    form.S3SessionToken,
		S3Bucket:          form.S3Bucket,
		S3Endpoint:        form.S3Endpoint,
		S3Region:          form.S3Region,
		S3UsePathStyle:    form.S3UsePathStyle,
	}
}

func emailSettingsFromForm(form emailSettingsForm) settingssvc.EmailInput {
	return settingssvc.EmailInput{
		Provider:      form.Provider,
		From:          form.From,
		PublicBaseURL: form.PublicBaseURL,
		SMTPHost:      form.SMTPHost,
		SMTPPort:      form.SMTPPort,
		SMTPUser:      form.SMTPUser,
		SMTPPass:      form.SMTPPass,
	}
}

func uploadSettingsFromForm(form uploadSettingsForm) settingssvc.UploadInput {
	return settingssvc.UploadInput{
		MaxQueueItems: form.MaxQueueItems,
	}
}

func renderSettingsStorage(c *gin.Context, user models.User, settings models.StorageSettings, validationErrors validation.Errors) {
	web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
		Ctx:             pages.ContextWithUser(user),
		StorageSettings: settings,
		StorageGB:       storageGB(settings.MaxStorageBytes),
		Errors:          validationErrors,
	}))
}

func renderSettingsEmail(c *gin.Context, user models.User, settings models.EmailSettings, validationErrors validation.Errors) {
	web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
		Ctx:           pages.ContextWithUser(user),
		EmailSettings: settings,
		Errors:        validationErrors,
	}))
}

func renderSettingsUpload(c *gin.Context, user models.User, settings models.UploadSettings, validationErrors validation.Errors) {
	web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
		Ctx:            pages.ContextWithUser(user),
		UploadSettings: settings,
		Errors:         validationErrors,
	}))
}

func storageGB(bytes int64) string {
	if bytes <= 0 || bytes == math.MaxInt64 {
		return "0"
	}
	return strconv.FormatInt(bytes/(1024*1024*1024), 10)
}
