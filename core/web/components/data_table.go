package components

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type DataTableProps struct {
	WrapClass  string
	TableClass string
	Header     g.Node
	Body       g.Node
	TopActions g.Node
	Pagination *PaginationProps
}

const DataTableCSS = "/web/components/data_table.css"

func DataTable(props DataTableProps) g.Node {
	wrapClass := "data-table-wrap"
	if props.WrapClass != "" {
		wrapClass += " " + props.WrapClass
	}

	tableClass := "data-table"
	if props.TableClass != "" {
		tableClass += " " + props.TableClass
	}

	var topToolbar g.Node
	if props.TopActions != nil || props.Pagination != nil {
		topToolbar = renderTableToolbar(props.TopActions, props.Pagination)
	}

	var bottomToolbar g.Node
	if props.Pagination != nil {
		bottomToolbar = renderTableToolbar(nil, props.Pagination)
	}

	return g.Group([]g.Node{
		InlineStyle(DataTableCSS),
		topToolbar,
		h.Div(
			h.Class(wrapClass),
			h.Table(
				h.Class(tableClass),
				g.If(props.Header != nil, props.Header),
				g.If(props.Body != nil, props.Body),
			),
		),
		bottomToolbar,
	})
}

func renderTableToolbar(actions g.Node, pagination *PaginationProps) g.Node {
	return h.Div(
		h.Class("data-table-toolbar"),
		g.If(actions != nil,
			h.Div(h.Class("data-table-toolbar-actions"), actions),
		),
		g.If(pagination != nil,
			h.Div(h.Class("data-table-toolbar-pagination"), Pagination(*pagination)),
		),
	)
}
