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

func WebPricing() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.PricingPage(pages.PricingPageProps{Ctx: pages.PageContext{}}))
	}
}

func WebContact() gin.HandlerFunc {
	return func(c *gin.Context) {
		web.Render(c, pages.ContactPage(pages.ContactPageProps{Ctx: pages.PageContext{}}))
	}
}


