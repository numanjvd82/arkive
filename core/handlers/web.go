package handlers

import (
	"arkive/core/web"
	pages "arkive/core/web/pages"

	"github.com/gin-gonic/gin"
)

func WebHome() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.HomePage(pages.HomePageProps{Ctx: pages.PageContext{}}))
	}
}

func WebPrivacy() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.PrivacyPage(pages.PrivacyPageProps{Ctx: pages.PageContext{}}))
	}
}

func WebCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.CookiePage(pages.CookiePageProps{Ctx: pages.PageContext{}}))
	}
}

func WebTerms() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.TermsPage(pages.TermsPageProps{Ctx: pages.PageContext{}}))
	}
}

func WebAUP() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.AUPPage(pages.AUPPageProps{Ctx: pages.PageContext{}}))
	}
}

func WebPricing() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.PricingPage(pages.PricingPageProps{Ctx: pages.PageContext{}}))
	}
}
