package components

import (
	"fmt"
	"net/url"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type PaginationProps struct {
	TotalRecords  int
	CurrentPage   int
	PageSize      int
	PageParam     string
	PageSizeParam string
	PageSizes     []int
	BaseURL       string
	Query         url.Values
}

const PaginationCSS = "/web/components/pagination.css"

func Pagination(props PaginationProps) g.Node {
	if props.TotalRecords <= 0 || props.PageSize <= 0 {
		return nil
	}

	pageParam := strings.TrimSpace(props.PageParam)
	if pageParam == "" {
		pageParam = "page"
	}
	pageSizeParam := strings.TrimSpace(props.PageSizeParam)
	if pageSizeParam == "" {
		pageSizeParam = "pageSize"
	}
	if len(props.PageSizes) == 0 {
		props.PageSizes = []int{25, 50, 100}
	}

	totalPages := (props.TotalRecords + props.PageSize - 1) / props.PageSize
	if totalPages <= 1 {
		return nil
	}
	if props.CurrentPage <= 0 {
		props.CurrentPage = 1
	}
	if props.CurrentPage > totalPages {
		props.CurrentPage = totalPages
	}

	pageLink := func(page int, label string, disabled bool) g.Node {
		classes := "pagination-link"
		if disabled {
			classes += " is-disabled"
		}
		linkQuery := cloneQuery(props.Query)
		linkQuery.Set(pageParam, fmt.Sprintf("%d", page))
		link := buildURL(props.BaseURL, linkQuery)
		if disabled {
			return h.Span(h.Class(classes), g.Text(label))
		}
		return h.A(h.Class(classes), h.Href(link), g.Text(label))
	}

	pages := []int{}
	start := props.CurrentPage - 2
	if start < 1 {
		start = 1
	}
	end := props.CurrentPage + 2
	if end > totalPages {
		end = totalPages
	}
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	return g.Group([]g.Node{
		InlineStyle(PaginationCSS),
		h.Div(
			h.Class("pagination"),
			h.Div(
				h.Class("pagination-pages"),
				pageLink(1, "First", props.CurrentPage == 1),
				pageLink(props.CurrentPage-1, "Prev", props.CurrentPage == 1),
				g.Group(g.Map(pages, func(page int) g.Node {
					label := fmt.Sprintf("%d", page)
					if page == props.CurrentPage {
						return h.Span(h.Class("pagination-link is-active"), g.Text(label))
					}
					return pageLink(page, label, false)
				})),
				pageLink(props.CurrentPage+1, "Next", props.CurrentPage == totalPages),
				pageLink(totalPages, "Last", props.CurrentPage == totalPages),
			),
			h.Select(
				h.Class("pagination-select"),
				g.Attr("onchange", "window.location=this.value"),
				g.Group(g.Map(props.PageSizes, func(size int) g.Node {
					selectQuery := cloneQuery(props.Query)
					selectQuery.Set(pageSizeParam, fmt.Sprintf("%d", size))
					selectQuery.Set(pageParam, "1")
					return h.Option(
						h.Value(buildURL(props.BaseURL, selectQuery)),
						g.Text(fmt.Sprintf("%d / page", size)),
						g.If(size == props.PageSize, h.Selected()),
					)
				})),
			),
		),
	})
}
