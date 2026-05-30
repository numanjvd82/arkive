package pages

import (
	"time"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"arkive/core/models"
	"arkive/core/web"
	"arkive/core/web/components"
)

type PublicShareViewProps struct {
	Token           string
	BurnAfterRead   bool
	File            models.File
	ViewURL         string
	DownloadURL     string
	IsImage         bool
	IsVideo         bool
	Viewable        bool
	ShareURL        string
	SharedAt        time.Time
	FileName        string
	MimeType        string
	PlaintextSize   int64
	PreviewText     string
	PreviewSettings models.PreviewSettings
}

func PublicShareViewPage(props PublicShareViewProps) web.Page {
	return web.Page{
		Title:    "Arkive · Shared file",
		Robots:   RobotsNoIndex,
		CSS:      []string{"/static/vendor/plyr/plyr.css", "/web/pages/share.css"},
		JS:       buildPublicShareJS(),
		ModuleJS: buildPublicShareModuleJS(),
		HideNav:  true,
		Body: g.Group([]g.Node{
			components.InlineStyle(components.UploadUICSS),
			h.Main(
				h.Class("public-share-page"),
				g.Attr("data-preview-image-max-bytes", int64String(props.PreviewSettings.ImageMaxBytes)),
				g.Attr("data-preview-video-max-bytes", int64String(props.PreviewSettings.VideoMaxBytes)),
				g.Attr("data-preview-text-max-bytes", int64String(props.PreviewSettings.TextMaxBytes)),
				h.Div(
					h.Class("public-share-shell"),
					renderPublicShareHeader(),
					renderPublicShareCard(props),
					renderPublicShareChecksum(props),
				),
				components.Lightbox(),
			),
			renderPublicShareFooter(),
		}),
	}
}
