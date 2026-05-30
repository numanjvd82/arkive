package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	filessvc "arkive/core/services/files"
	settingssvc "arkive/core/services/settings"
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

type uploadSettingsForm struct {
	MaxQueueItems    string `form:"max_queue_items"`
	PartConcurrency  string `form:"part_concurrency"`
	StaleUploadHours string `form:"stale_upload_hours"`
}

type previewSettingsForm struct {
	ImageMaxMB string `form:"image_max_mb"`
	VideoMaxMB string `form:"video_max_mb"`
	TextMaxMB  string `form:"text_max_mb"`
}

func WebSettings(filesService *filessvc.Service, settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err := filesService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		storageSettings, _ := settingsService.StorageSettings(c.Request.Context())
		uploadSettings, _ := settingsService.UploadSettings(c.Request.Context())
		previewSettings, _ := settingsService.PreviewSettings(c.Request.Context())
		web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
			Ctx:             pages.ContextWithUser(user),
			StorageSettings: storageSettings,
			UploadSettings:  uploadSettings,
			PreviewSettings: previewSettings,
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

func WebSettingsPreviewPost(settingsService *settingssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var form previewSettingsForm
		if err := c.ShouldBind(&form); err != nil {
			renderSettingsPreview(c, user, models.PreviewSettings{}, validation.Errors{validation.GeneralKey: "Please fill out the form."})
			return
		}

		settings, validationErrors, err := settingsService.UpdatePreviewSettings(c.Request.Context(), user.ID, previewSettingsFromForm(form))
		if validationErrors != nil && validationErrors.HasAny() {
			renderSettingsPreview(c, user, settings, validationErrors)
			return
		}
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Redirect(http.StatusSeeOther, "/settings?msg=preview-updated")
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

func uploadSettingsFromForm(form uploadSettingsForm) settingssvc.UploadInput {
	return settingssvc.UploadInput{
		MaxQueueItems:    form.MaxQueueItems,
		PartConcurrency:  form.PartConcurrency,
		StaleUploadHours: form.StaleUploadHours,
	}
}

func previewSettingsFromForm(form previewSettingsForm) settingssvc.PreviewInput {
	return settingssvc.PreviewInput{
		ImageMaxMB: form.ImageMaxMB,
		VideoMaxMB: form.VideoMaxMB,
		TextMaxMB:  form.TextMaxMB,
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

func renderSettingsUpload(c *gin.Context, user models.User, settings models.UploadSettings, validationErrors validation.Errors) {
	web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
		Ctx:            pages.ContextWithUser(user),
		UploadSettings: settings,
		Errors:         validationErrors,
	}))
}

func renderSettingsPreview(c *gin.Context, user models.User, settings models.PreviewSettings, validationErrors validation.Errors) {
	web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
		Ctx:             pages.ContextWithUser(user),
		PreviewSettings: settings,
		Errors:          validationErrors,
	}))
}

func storageGB(bytes int64) string {
	if bytes <= 0 || bytes == math.MaxInt64 {
		return "0"
	}
	return strconv.FormatInt(bytes/(1024*1024*1024), 10)
}
