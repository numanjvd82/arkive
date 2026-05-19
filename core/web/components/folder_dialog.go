package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func FolderDialog() g.Node {
	return Dialog(DialogProps{
		BackdropID: "folder-create-backdrop",
		TitleID:    "folder-create-title",
		Title:      "New folder",
		Body: h.Div(
			h.Class("folder-dialog"),
			h.Label(
				h.Class("form-label"),
				g.Attr("for", "folder-name-input"),
				g.Text("Folder name"),
			),
			h.Input(
				h.Class("form-input"),
				g.Attr("id", "folder-name-input"),
				g.Attr("type", "text"),
				g.Attr("maxlength", "255"),
				g.Attr("placeholder", "Photos"),
			),
		),
		Actions: g.Group([]g.Node{
			h.Button(
				h.Class("button secondary"),
				h.Type("button"),
				g.Attr("id", "folder-create-cancel"),
				g.Text("Cancel"),
			),
			h.Button(
				h.Class("button primary"),
				h.Type("button"),
				g.Attr("id", "folder-create-confirm"),
				g.Text("Create folder"),
			),
		}),
	})
}
