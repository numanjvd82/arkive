package components

import (
	"strings"

	lucide "github.com/eduardolat/gomponents-lucide"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type ButtonProps struct {
	Text     string
	Variant  string
	Type     string
	Href     string
	Class    string
	ID       string
	Icon     string
	BusyText string
}

const ButtonCSS = "/web/components/button.css"

func Button(props ButtonProps) g.Node {
	classes := []string{"button"}
	if props.Variant != "" {
		classes = append(classes, strings.TrimSpace(props.Variant))
	}
	if props.Class != "" {
		classes = append(classes, strings.TrimSpace(props.Class))
	}
	className := strings.Join(classes, " ")

	if props.Href != "" {
		return g.Group([]g.Node{
			InlineStyle(ButtonCSS),
			h.A(
				h.Class(className),
				h.Href(props.Href),
				g.If(props.ID != "", g.Attr("id", props.ID)),
				g.If(props.Icon != "", h.Span(
					h.Class("button-icon"),
					renderButtonIcon(props.Icon),
				)),
				h.Span(h.Class("button-label"), g.Text(props.Text)),
			),
		})
	}

	buttonType := props.Type
	if buttonType == "" {
		buttonType = "button"
	}

	return g.Group([]g.Node{
		InlineStyle(ButtonCSS),
		h.Button(
			h.Class(className),
			g.Attr("type", buttonType),
			g.If(props.ID != "", g.Attr("id", props.ID)),
			g.If(props.BusyText != "", g.Attr("data-busy-text", props.BusyText)),
			g.If(props.Icon != "", h.Span(
				h.Class("button-icon"),
				renderButtonIcon(props.Icon),
			)),
			h.Span(h.Class("button-label"), g.Text(props.Text)),
		),
	})
}

func renderButtonIcon(name string) g.Node {
	switch name {
	case "copy":
		return lucide.Copy(
			h.Class("button-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	case "key":
		return lucide.Key(
			h.Class("button-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	case "download":
		return lucide.Download(
			h.Class("button-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	case "share":
		return lucide.Share2(
			h.Class("button-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	case "trash":
		return lucide.Trash2(
			h.Class("button-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	case "arrow-left":
		return lucide.ArrowLeft(
			h.Class("button-lucide"),
			g.Attr("aria-hidden", "true"),
		)
	default:
		return Icon(IconProps{
			Name:       name,
			Size:       "18",
			Decorative: true,
		})
	}
}
