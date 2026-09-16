package render

import (
	"strings"
	"testing"

	"github.com/BergurHeimisson/md-viewer/internal/ansi"
)

// renderPlain renders without wrapping and strips colour, for assertions about
// structure rather than styling.
func renderPlain(t *testing.T, md string) string {
	t.Helper()
	r := &Renderer{}
	return ansi.Strip(r.Render([]byte(md)))
}

func renderColoured(t *testing.T, md string) string {
	t.Helper()
	r := &Renderer{}
	return r.Render([]byte(md))
}

func TestHeadings(t *testing.T) {
	tests := []struct {
		name   string
		md     string
		text   string
		colour string
	}{
		{"h1", "# Title", "═══ Title", ansi.CyanBold},
		{"h2", "## Section", "─── Section", ansi.YellowBoldBright},
		{"h3", "### Sub", "▸ Sub", ansi.PurpleBold},
		{"h4 uses the h3 style", "#### Deeper", "▸ Deeper", ansi.PurpleBold},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderPlain(t, tt.md); !strings.Contains(got, tt.text) {
				t.Errorf("got %q, want it to contain %q", got, tt.text)
			}
			if got := renderColoured(t, tt.md); !strings.HasPrefix(got, tt.colour) {
				t.Errorf("got %q, want prefix %q", got, tt.colour)
			}
		})
	}
}

func TestEmphasis(t *testing.T) {
	tests := []struct {
		name   string
		md     string
		colour string
	}{
		{"strong", "**bold**", ansi.WhiteBold},
		{"emphasis", "*italic*", ansi.Yellow},
		{"code span", "`code`", ansi.Cyan},
		{"strikethrough", "~~gone~~", ansi.BlackBright},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderColoured(t, tt.md)
			if !strings.Contains(got, tt.colour) {
				t.Errorf("got %q, want it to contain colour %q", got, tt.colour)
			}
			if !strings.Contains(got, ansi.Reset) {
				t.Errorf("got %q, want a reset", got)
			}
		})
	}
}

func TestCodeBlockGetsAGutterPerLine(t *testing.T) {
	got := renderPlain(t, "```\nfirst\nsecond\n```")
	want := "  │ first\n  │ second\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIndentedCodeBlock(t *testing.T) {
	got := renderPlain(t, "    indented code\n")
	if want := "  │ indented code\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBulletList(t *testing.T) {
	got := renderPlain(t, "- one\n- two\n")
	want := "  • one\n  • two\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOrderedListHonoursStartNumber(t *testing.T) {
	got := renderPlain(t, "3. three\n4. four\n")
	want := "  3. three\n  4. four\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNestedListIsIndented(t *testing.T) {
	got := renderPlain(t, "- outer\n  - inner\n")
	want := "  • outer\n      • inner\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTaskListReplacesTheBulletWithACheckbox(t *testing.T) {
	got := renderPlain(t, "- [x] done\n- [ ] todo\n")
	want := "  ☑ done\n  ☐ todo\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	coloured := renderColoured(t, "- [x] done\n")
	if !strings.Contains(coloured, ansi.GreenBold+"☑") {
		t.Errorf("checked box should be green bold, got %q", coloured)
	}
}

func TestBlockquotePrefixesEveryLine(t *testing.T) {
	got := renderPlain(t, "> first\n>\n> second\n")
	want := "▌ first\n▌ \n▌ second\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLinkShowsItsDestination(t *testing.T) {
	got := renderPlain(t, "[docs](https://example.com)")
	if want := "docs [https://example.com]\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAutolinkedBareURL(t *testing.T) {
	got := renderPlain(t, "see https://example.com now")
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("got %q, want it to contain the URL", got)
	}
}

func TestImageShowsAltTextAndDestination(t *testing.T) {
	got := renderPlain(t, "![a cat](cat.png)")
	if want := "🖼 a cat [cat.png]\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestThematicBreak(t *testing.T) {
	got := renderPlain(t, "---\n")
	if want := strings.Repeat("─", 60) + "\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Raw markup in prose is preserved literally, so placeholders like <min> in
// documentation survive instead of vanishing.
func TestInlineHTMLIsPreserved(t *testing.T) {
	got := renderPlain(t, "range <min> to <max>")
	if want := "range <min> to <max>\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHTMLBlockIsPreserved(t *testing.T) {
	got := renderPlain(t, "<div>\nraw\n</div>\n")
	if !strings.Contains(got, "<div>") || !strings.Contains(got, "raw") {
		t.Errorf("got %q, want the raw block preserved", got)
	}
}

func TestBlocksAreSeparatedByABlankLine(t *testing.T) {
	got := renderPlain(t, "one\n\ntwo\n")
	if want := "one\n\ntwo\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSoftLineBreakBecomesASpace(t *testing.T) {
	got := renderPlain(t, "first\nsecond\n")
	if want := "first second\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHardLineBreakIsKept(t *testing.T) {
	got := renderPlain(t, "first  \nsecond\n")
	if want := "first\nsecond\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmptyInputRendersNothing(t *testing.T) {
	if got := renderPlain(t, ""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestWrappingRespectsTheConfiguredWidth(t *testing.T) {
	const width = 24
	r := &Renderer{Width: width}
	md := "The quick brown fox jumps over the lazy dog and then keeps on running."
	for _, line := range strings.Split(ansi.Strip(r.Render([]byte(md))), "\n") {
		if n := len([]rune(line)); n > width {
			t.Errorf("line %q is %d cells, want <= %d", line, n, width)
		}
	}
}

// Code blocks are pre-formatted; re-flowing them would destroy their meaning.
func TestWrappingLeavesCodeBlocksAlone(t *testing.T) {
	r := &Renderer{Width: 20}
	md := "```\nthis line is far longer than twenty columns\n```"
	got := ansi.Strip(r.Render([]byte(md)))
	if want := "  │ this line is far longer than twenty columns\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The heading markers are multi-byte ("═══ " is 10 bytes, 4 cells). Measuring
// them with len() wrapped headings too narrow and over-indented their
// continuation lines.
func TestWrappedHeadingLinesUpUnderItsText(t *testing.T) {
	const width = 30
	r := &Renderer{Width: width}
	got := ansi.Strip(r.Render([]byte("# alpha beta gamma delta epsilon zeta eta")))

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected the heading to wrap, got %q", got)
	}
	for _, line := range lines {
		if n := len([]rune(line)); n > width {
			t.Errorf("line %q is %d cells, want <= %d", line, n, width)
		}
	}
	for _, line := range lines[1:] {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent != 4 { // len("═══ ") in cells
			t.Errorf("continuation %q indented %d, want 4", line, indent)
		}
	}
}

func TestHeadingWrapsAtNarrowWidths(t *testing.T) {
	// width-len(marker) used to go negative here, silently disabling wrapping.
	r := &Renderer{Width: 8}
	got := ansi.Strip(r.Render([]byte("# alpha bravo charlie")))
	if !strings.Contains(got, "\n") {
		t.Errorf("heading should still wrap at width 8, got %q", got)
	}
}

// CommonMark allows an ordered list to start at 0.
func TestOrderedListCanStartAtZero(t *testing.T) {
	got := renderPlain(t, "0. zero\n1. one\n")
	if want := "  0. zero\n  1. one\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
