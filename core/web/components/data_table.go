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

	return g.Group([]g.Node{
		InlineStyle(DataTableCSS),
		g.If(props.TopActions != nil || props.Pagination != nil,
			renderTableToolbar(props.TopActions, props.Pagination),
		),
		h.Div(
			h.Class(wrapClass),
			h.Table(
				h.Class(tableClass),
				g.If(props.Header != nil, props.Header),
				g.If(props.Body != nil, props.Body),
			),
		),
		g.If(props.Pagination != nil,
			renderTableToolbar(nil, props.Pagination),
		),
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
