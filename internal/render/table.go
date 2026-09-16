package render

import (
	"strings"

	"github.com/BergurHeimisson/md-viewer/internal/ansi"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// table renders a GFM table as a box-drawn grid with columns padded to their
// widest cell. Tables are never wrapped; an over-wide table is clipped by the
// terminal, which keeps the column alignment intact.
func (w *walker) table(n *extast.Table) string {
	var (
		rows      [][]string
		aligns    []extast.Alignment
		headerLen int
	)

	// goldmark hangs the header and every body row directly off the table;
	// there is no body wrapper node.
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		switch row.(type) {
		case *extast.TableHeader:
			rows = append(rows, w.tableRow(row, &aligns))
			headerLen++
		case *extast.TableRow:
			rows = append(rows, w.tableRow(row, &aligns))
		}
	}

	if len(rows) == 0 {
		return ""
	}

	columns := 0
	for _, row := range rows {
		if len(row) > columns {
			columns = len(row)
		}
	}
	widths := make([]int, columns)
	for _, row := range rows {
		for c, cell := range row {
			if n := ansi.VisibleLen(cell); n > widths[c] {
				widths[c] = n
			}
		}
	}

	var b strings.Builder
	b.WriteString(border(widths, "┌", "┬", "┐"))
	for i, row := range rows {
		b.WriteString(w.tableLine(row, widths, aligns, i < headerLen))
		if i == headerLen-1 {
			b.WriteString(border(widths, "├", "┼", "┤"))
		}
	}
	b.WriteString(border(widths, "└", "┴", "┘"))
	return b.String()
}

func (w *walker) tableRow(row ast.Node, aligns *[]extast.Alignment) []string {
	var cells []string
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		tc, ok := cell.(*extast.TableCell)
		if !ok {
			continue
		}
		if len(*aligns) < len(cells)+1 {
			*aligns = append(*aligns, tc.Alignment)
		}
		cells = append(cells, strings.TrimSpace(w.inline(tc)))
	}
	return cells
}

func border(widths []int, left, mid, right string) string {
	var b strings.Builder
	b.WriteString(ansi.BlackBright)
	b.WriteString(left)
	for i, width := range widths {
		b.WriteString(strings.Repeat("─", width+2))
		if i == len(widths)-1 {
			b.WriteString(right)
		} else {
			b.WriteString(mid)
		}
	}
	b.WriteString(ansi.Reset)
	b.WriteByte('\n')
	return b.String()
}

func (w *walker) tableLine(cells []string, widths []int, aligns []extast.Alignment, header bool) string {
	var b strings.Builder
	for c, width := range widths {
		b.WriteString(ansi.BlackBright + "│" + ansi.Reset + " ")

		content := ""
		if c < len(cells) {
			content = cells[c]
		}
		if header {
			content = ansi.WhiteBold + content + ansi.Reset
		}

		gap := width - ansi.VisibleLen(content)
		align := extast.AlignNone
		if c < len(aligns) {
			align = aligns[c]
		}
		left, right := pad(gap, align)
		b.WriteString(left + content + right + " ")
	}
	b.WriteString(ansi.BlackBright + "│" + ansi.Reset + "\n")
	return b.String()
}

// pad splits gap spaces either side of a cell according to its alignment.
func pad(gap int, align extast.Alignment) (string, string) {
	if gap < 0 {
		gap = 0
	}
	switch align {
	case extast.AlignRight:
		return strings.Repeat(" ", gap), ""
	case extast.AlignCenter:
		l := gap / 2
		return strings.Repeat(" ", l), strings.Repeat(" ", gap-l)
	default:
		return "", strings.Repeat(" ", gap)
	}
}
