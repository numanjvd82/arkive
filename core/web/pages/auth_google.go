package pages

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func googleAuthSection(clientID string) g.Node {
	if clientID == "" {
		return nil
	}

	return h.Div(
		h.Class("auth-oauth"),
		g.Attr("data-google-client-id", clientID),
		g.Attr("data-google-login-endpoint", "/auth/google"),
		h.Div(
			h.Class("auth-oauth-button"),
			h.ID("google-signin-button"),
		),
	)
}
