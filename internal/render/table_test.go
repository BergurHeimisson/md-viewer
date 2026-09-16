package render

import (
	"strings"
	"testing"

	"github.com/BergurHeimisson/md-viewer/internal/ansi"
)

const simpleTable = `| Key | Action |
|---|---|
| q | quit |
| SPACE | next screen |
`

func TestTableIsDrawnAsAGrid(t *testing.T) {
	got := renderPlain(t, simpleTable)
	want := strings.Join([]string{
		"┌───────┬─────────────┐",
		"│ Key   │ Action      │",
		"├───────┼─────────────┤",
		"│ q     │ quit        │",
		"│ SPACE │ next screen │",
		"└───────┴─────────────┘",
		"",
	}, "\n")
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestTableHeaderIsBold(t *testing.T) {
	got := renderColoured(t, simpleTable)
	if !strings.Contains(got, ansi.WhiteBold+"Key"+ansi.Reset) {
		t.Errorf("header cell should be bold, got:\n%s", got)
	}
}

// Column widths are measured on visible text, so emphasis inside a cell must
// not push the grid out of alignment.
func TestTableAlignsAroundInlineStyling(t *testing.T) {
	md := "| A | B |\n|---|---|\n| **bold** | x |\n"
	lines := strings.Split(strings.TrimRight(renderColoured(t, md), "\n"), "\n")
	width := ansi.VisibleLen(lines[0])
	for _, line := range lines {
		if n := ansi.VisibleLen(line); n != width {
			t.Errorf("line %q has visible width %d, want %d", ansi.Strip(line), n, width)
		}
	}
}

func TestTableRespectsColumnAlignment(t *testing.T) {
	md := "| Left | Center | Right |\n|:--|:-:|--:|\n| a | b | c |\n"
	lines := strings.Split(strings.TrimRight(renderPlain(t, md), "\n"), "\n")
	dataRow := lines[3]

	cells := strings.Split(strings.Trim(dataRow, "│"), "│")
	if len(cells) != 3 {
		t.Fatalf("expected 3 cells in %q, got %d", dataRow, len(cells))
	}

	tests := []struct {
		name               string
		cell               string
		wantLead, wantTail int
	}{
		{"left", cells[0], 1, 4},
		{"center", cells[1], 3, 4},
		{"right", cells[2], 5, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lead := len(tt.cell) - len(strings.TrimLeft(tt.cell, " "))
			tail := len(tt.cell) - len(strings.TrimRight(tt.cell, " "))
			if lead != tt.wantLead || tail != tt.wantTail {
				t.Errorf("cell %q padded %d/%d, want %d/%d", tt.cell, lead, tail, tt.wantLead, tt.wantTail)
			}
		})
	}
}

func TestTableWithNoBodyRows(t *testing.T) {
	got := renderPlain(t, "| Only |\n|---|\n")
	want := strings.Join([]string{
		"┌──────┐",
		"│ Only │",
		"├──────┤",
		"└──────┘",
		"",
	}, "\n")
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestTableWithAShortRow(t *testing.T) {
	md := "| A | B |\n|---|---|\n| x |\n"
	lines := strings.Split(strings.TrimRight(renderPlain(t, md), "\n"), "\n")
	width := len([]rune(lines[0]))
	for _, line := range lines {
		if n := len([]rune(line)); n != width {
			t.Errorf("line %q has width %d, want %d", line, n, width)
		}
	}
}
