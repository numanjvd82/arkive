package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func MoveDialog() g.Node {
	return Dialog(DialogProps{
		BackdropID: "entries-move-backdrop",
		TitleID:    "entries-move-title",
		Title:      "Move entries",
		Body: h.Div(
			h.Class("move-dialog"),
			h.P(g.Attr("id", "move-entries-meta"), g.Text("Choose a destination folder.")),
			h.Label(
				h.Class("form-label"),
				g.Attr("for", "move-target-folder"),
				g.Text("Destination"),
			),
			h.Select(
				h.Class("form-input"),
				g.Attr("id", "move-target-folder"),
			),
		),
		Actions: g.Group([]g.Node{
			h.Button(
				h.Class("button secondary"),
				h.Type("button"),
				g.Attr("id", "move-entries-cancel"),
				g.Text("Cancel"),
			),
			h.Button(
				h.Class("button primary"),
				h.Type("button"),
				g.Attr("id", "move-entries-confirm"),
				g.Text("Move"),
			),
		}),
	})
}
