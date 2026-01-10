package pages

import "arkive/core/models"

type PageContext struct {
	User *models.User
}

const (
	DefaultOGImage = "/static/assets/images/og-image.png"
	RobotsIndex    = "index, follow"
	RobotsNoIndex  = "noindex, nofollow"
)

func ContextWithUser(user models.User) PageContext {
	return PageContext{User: &user}
}
