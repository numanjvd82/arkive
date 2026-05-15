package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	g "maragu.dev/gomponents"

	"arkive/core/models"
)

type Page struct {
	Title              string
	Description        string
	CanonicalURL       string
	CanonicalPath      string
	Robots             string
	OGTitle            string
	OGDescription      string
	OGImage            string
	OGType             string
	TwitterCard        string
	JSONLD             string
	CSS                []string
	JS                 []string
	ModuleJS           []string
	Body               g.Node
	HideNav            bool
	AuthLayout         bool
	RequireVaultUnlock bool
	User               *models.User
	ActiveNav          string
	SearchPlaceholder  string
}

func Render(c *gin.Context, page Page) {
	baseURL := buildBaseURL(c)
	canonicalURL := page.CanonicalURL
	if canonicalURL == "" && page.CanonicalPath != "" {
		canonicalURL = buildAbsoluteURL(baseURL, page.CanonicalPath)
	}
	ogTitle := page.OGTitle
	if ogTitle == "" {
		ogTitle = page.Title
	}
	ogDescription := page.OGDescription
	if ogDescription == "" {
		ogDescription = page.Description
	}
	ogType := page.OGType
	if ogType == "" {
		ogType = "website"
	}
	ogImage := buildAbsoluteURL(baseURL, page.OGImage)
	twitterCard := page.TwitterCard
	if twitterCard == "" {
		if ogImage != "" {
			twitterCard = "summary_large_image"
		} else {
			twitterCard = "summary"
		}
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	var node g.Node
	if page.AuthLayout {
		if page.Body == nil {
			node = AuthLayout(LayoutData{
				Title:              page.Title,
				Description:        page.Description,
				CanonicalURL:       canonicalURL,
				Robots:             page.Robots,
				OGTitle:            ogTitle,
				OGDescription:      ogDescription,
				OGImage:            ogImage,
				OGType:             ogType,
				TwitterCard:        twitterCard,
				JSONLD:             page.JSONLD,
				CSS:                page.CSS,
				JS:                 page.JS,
				ModuleJS:           page.ModuleJS,
				User:               page.User,
				ActiveNav:          page.ActiveNav,
				SearchPlaceholder:  page.SearchPlaceholder,
				RequireVaultUnlock: page.RequireVaultUnlock,
			})
		} else {
			node = AuthLayout(LayoutData{
				Title:              page.Title,
				Description:        page.Description,
				CanonicalURL:       canonicalURL,
				Robots:             page.Robots,
				OGTitle:            ogTitle,
				OGDescription:      ogDescription,
				OGImage:            ogImage,
				OGType:             ogType,
				TwitterCard:        twitterCard,
				JSONLD:             page.JSONLD,
				CSS:                page.CSS,
				JS:                 page.JS,
				ModuleJS:           page.ModuleJS,
				User:               page.User,
				ActiveNav:          page.ActiveNav,
				SearchPlaceholder:  page.SearchPlaceholder,
				RequireVaultUnlock: page.RequireVaultUnlock,
			}, page.Body)
		}
	} else if page.Body == nil {
		node = Layout(LayoutData{
			Title:         page.Title,
			Description:   page.Description,
			CanonicalURL:  canonicalURL,
			Robots:        page.Robots,
			OGTitle:       ogTitle,
			OGDescription: ogDescription,
			OGImage:       ogImage,
			OGType:        ogType,
			TwitterCard:   twitterCard,
			JSONLD:        page.JSONLD,
			CSS:           page.CSS,
			JS:            page.JS,
			ModuleJS:      page.ModuleJS,
			HideNav:       page.HideNav,
			User:          page.User,
		})
	} else {
		node = Layout(LayoutData{
			Title:         page.Title,
			Description:   page.Description,
			CanonicalURL:  canonicalURL,
			Robots:        page.Robots,
			OGTitle:       ogTitle,
			OGDescription: ogDescription,
			OGImage:       ogImage,
			OGType:        ogType,
			TwitterCard:   twitterCard,
			JSONLD:        page.JSONLD,
			CSS:           page.CSS,
			JS:            page.JS,
			ModuleJS:      page.ModuleJS,
			HideNav:       page.HideNav,
			User:          page.User,
		}, page.Body)
	}
	if err := node.Render(c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

func buildBaseURL(c *gin.Context) string {
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return ""
	}

	return scheme + "://" + host
}

func buildAbsoluteURL(baseURL, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	if strings.HasPrefix(value, "/") && baseURL != "" {
		return baseURL + value
	}
	return value
}
