package components

import (
	"net/url"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type SortOption struct {
	Label string
	Value string
}

type SortSelectProps struct {
	ID       string
	Label    string
	Value    string
	Options  []SortOption
	QueryKey string
	BaseURL  string
	Query    url.Values
	Class    string
}

const SortSelectCSS = "/web/components/sort.css"

func SortSelect(props SortSelectProps) g.Node {
	if len(props.Options) == 0 {
		return nil
	}
	queryKey := strings.TrimSpace(props.QueryKey)
	if queryKey == "" {
		queryKey = "sort"
	}

	selectID := strings.TrimSpace(props.ID)
	if selectID == "" {
		selectID = "sort-select"
	}

	classes := "sort-select"
	if strings.TrimSpace(props.Class) != "" {
		classes += " " + strings.TrimSpace(props.Class)
	}

	return g.Group([]g.Node{
		InlineStyle(SortSelectCSS),
		h.Label(
			h.Class("sort-control"),
			g.If(strings.TrimSpace(props.Label) != "", h.Span(h.Class("sort-label"), g.Text(props.Label))),
			h.Select(
				h.ID(selectID),
				h.Class(classes),
				g.Attr("onchange", "window.location=this.value"),
				g.Group(g.Map(props.Options, func(option SortOption) g.Node {
					optionQuery := cloneQuery(props.Query)
					optionQuery.Set(queryKey, option.Value)
					urlValue := buildURL(props.BaseURL, optionQuery)
					return h.Option(
						h.Value(urlValue),
						g.Text(option.Label),
						g.If(option.Value == props.Value, h.Selected()),
					)
				})),
			),
		),
	})
}

func cloneQuery(values url.Values) url.Values {
	clone := url.Values{}
	for key, items := range values {
		copied := make([]string, len(items))
		copy(copied, items)
		clone[key] = copied
	}
	return clone
}

func buildURL(baseURL string, query url.Values) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = ""
	}
	encoded := query.Encode()
	if encoded == "" {
		return base
	}
	return base + "?" + encoded
}
