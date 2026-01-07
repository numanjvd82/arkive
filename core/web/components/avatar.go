package components

import (
	"fmt"
	"strings"
	"unicode/utf8"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type AvatarProps struct {
	Text       string
	Size       int
	Class      string
	Title      string
	Decorative bool
}

const AvatarCSS = "/web/components/avatar.css"

func Avatar(props AvatarProps) g.Node {
	trimmed := strings.TrimSpace(props.Text)
	initial := "?"
	if trimmed != "" {
		r, _ := utf8.DecodeRuneInString(trimmed)
		initial = strings.ToUpper(string(r))
	}

	classes := "avatar"
	if strings.TrimSpace(props.Class) != "" {
		classes = classes + " " + strings.TrimSpace(props.Class)
	}

	nodes := []g.Node{
		h.Class(classes),
		g.Text(initial),
	}
	if props.Size > 0 {
		nodes = append(nodes, g.Attr("style", fmt.Sprintf("--avatar-size: %dpx", props.Size)))
	}
	if props.Decorative {
		nodes = append(nodes, g.Attr("aria-hidden", "true"))
	} else if props.Title != "" {
		nodes = append(nodes, g.Attr("role", "img"), g.Attr("aria-label", props.Title))
	} else if trimmed != "" {
		nodes = append(nodes, g.Attr("role", "img"), g.Attr("aria-label", trimmed))
	}

	return g.Group([]g.Node{
		InlineStyle(AvatarCSS),
		h.Span(nodes...),
	})
}
