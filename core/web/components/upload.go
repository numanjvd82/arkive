package components

import (
	"strconv"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type UploadInputProps struct {
	ID       string
	Name     string
	Label    string
	Helper   string
	Accept   string
	Required bool
}

func UploadInput(props UploadInputProps) g.Node {
	id := props.ID
	if id == "" {
		id = props.Name
	}
	input := h.Input(
		h.Type("file"),
		h.ID(id),
		h.Name(props.Name),
		h.Class("form-input upload-input"),
		g.If(props.Accept != "", g.Attr("accept", props.Accept)),
		g.If(props.Required, g.Attr("required", "required")),
	)

	helperID := ""
	if props.Helper != "" {
		helperID = id + "-helper"
		input = h.Input(
			h.Type("file"),
			h.ID(id),
			h.Name(props.Name),
			h.Class("form-input upload-input"),
			g.If(props.Accept != "", g.Attr("accept", props.Accept)),
			g.If(props.Required, g.Attr("required", "required")),
			g.Attr("aria-describedby", helperID),
		)
	}

	return h.Div(
		h.Class("form-field"),
		h.Label(
			h.Class("form-label"),
			h.For(id),
			g.Text(props.Label),
		),
		input,
		g.If(props.Helper != "",
			h.P(
				h.Class("form-subtitle"),
				h.ID(helperID),
				g.Text(props.Helper),
			),
		),
	)
}

type ProgressBarProps struct {
	ID    string
	Value int
	Label string
}

func ProgressBar(props ProgressBarProps) g.Node {
	value := props.Value
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}

	return h.Div(
		h.Class("progress"),
		g.If(props.ID != "", g.Attr("id", props.ID)),
		g.If(props.Label != "",
			h.Span(h.Class("progress-label"), g.Text(props.Label)),
		),
		h.Div(
			h.Class("progress-track"),
			h.Div(
				h.Class("progress-bar"),
				g.Attr("style", "width: "+strconv.Itoa(value)+"%"),
				g.Attr("role", "progressbar"),
				g.Attr("aria-valuemin", "0"),
				g.Attr("aria-valuemax", "100"),
				g.Attr("aria-valuenow", strconv.Itoa(value)),
			),
		),
	)
}

type UploadControlsProps struct {
	InputID       string
	InputName     string
	InputLabel    string
	InputHelper   string
	StatusText    string
	StartLabel    string
	InputRequired bool
}

func UploadControls(props UploadControlsProps) g.Node {
	inputID := props.InputID
	if inputID == "" {
		inputID = "upload-file"
	}
	inputName := props.InputName
	if inputName == "" {
		inputName = "file"
	}
	startLabel := props.StartLabel
	if startLabel == "" {
		startLabel = "Start upload"
	}
	statusText := props.StatusText
	if statusText == "" {
		statusText = "No uploads yet."
	}

	return g.Group([]g.Node{
		UploadInput(UploadInputProps{
			ID:       inputID,
			Name:     inputName,
			Label:    props.InputLabel,
			Helper:   props.InputHelper,
			Required: props.InputRequired,
		}),
		h.Div(
			h.Class("upload-actions"),
			h.Button(
				h.Class("button primary"),
				h.Type("button"),
				g.Attr("id", "upload-start"),
				g.Text(startLabel),
			),
			h.Button(
				h.Class("icon-button"),
				h.Type("button"),
				g.Attr("id", "upload-pause"),
				g.Attr("disabled", "disabled"),
				Icon(IconProps{
					Name:       "pause",
					Size:       "md",
					Title:      "Pause upload",
					AriaLabel:  "Pause upload",
					Decorative: false,
				}),
			),
			h.Button(
				h.Class("icon-button"),
				h.Type("button"),
				g.Attr("id", "upload-resume"),
				g.Attr("disabled", "disabled"),
				Icon(IconProps{
					Name:       "play",
					Size:       "md",
					Title:      "Resume upload",
					AriaLabel:  "Resume upload",
					Decorative: false,
				}),
			),
			h.Button(
				h.Class("icon-button"),
				h.Type("button"),
				g.Attr("id", "upload-abort"),
				g.Attr("disabled", "disabled"),
				Icon(IconProps{
					Name:       "x",
					Size:       "md",
					Title:      "Abort upload",
					AriaLabel:  "Abort upload",
					Decorative: false,
				}),
			),
		),
		ProgressBar(ProgressBarProps{
			ID:    "upload-progress",
			Value: 0,
			Label: "Progress",
		}),
		h.P(
			h.Class("upload-status"),
			g.Attr("id", "upload-status"),
			g.Text(statusText),
		),
	})
}
