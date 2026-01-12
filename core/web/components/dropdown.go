package components

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type DropdownProps struct {
	ID        string
	Align     string
	Label     string
	Trigger   g.Node
	Menu      g.Node
	Class     string
	MenuClass string
}

const DropdownCSS = "/web/components/dropdown.css"
const DropdownJS = "/web/components/dropdown.js"

func Dropdown(props DropdownProps) g.Node {
	id := strings.TrimSpace(props.ID)
	menuID := ""
	if id != "" {
		menuID = id + "-menu"
	}

	alignClass := "is-right"
	if strings.EqualFold(props.Align, "left") {
		alignClass = "is-left"
	}

	rootClasses := "dropdown " + alignClass
	if strings.TrimSpace(props.Class) != "" {
		rootClasses += " " + strings.TrimSpace(props.Class)
	}

	menuClasses := "dropdown-menu"
	if strings.TrimSpace(props.MenuClass) != "" {
		menuClasses += " " + strings.TrimSpace(props.MenuClass)
	}

	triggerAttrs := []g.Node{
		h.Type("button"),
		h.Class("dropdown-trigger"),
		g.Attr("data-dropdown-trigger", "true"),
		g.Attr("aria-haspopup", "menu"),
		g.Attr("aria-expanded", "false"),
	}
	if menuID != "" {
		triggerAttrs = append(triggerAttrs, g.Attr("aria-controls", menuID))
	}
	if strings.TrimSpace(props.Label) != "" {
		triggerAttrs = append(triggerAttrs, g.Attr("aria-label", strings.TrimSpace(props.Label)))
	}

	return g.Group([]g.Node{
		InlineStyle(DropdownCSS),
		InlineScript(DropdownJS),
		h.Div(
			h.Class(rootClasses),
			g.Attr("data-dropdown", "true"),
			h.Button(
				g.Group(triggerAttrs),
				props.Trigger,
			),
			h.Div(
				g.If(menuID != "", h.ID(menuID)),
				h.Class(menuClasses),
				g.Attr("data-dropdown-menu", "true"),
				g.Attr("role", "menu"),
				props.Menu,
			),
		),
	})
}
