package pages

import "arkive/core/models"

type PageContext struct {
	User *models.User
}

const (
	DefaultOGImage = "/static/assets/images/android-chrome-512x512.png"
	RobotsIndex    = "index, follow"
	RobotsNoIndex  = "noindex, nofollow"
)

func ContextWithUser(user models.User) PageContext {
	return PageContext{User: &user}
}
