// Package render turns a Markdown document into ANSI-coloured terminal text.
package render

import (
	"strconv"
	"strings"

	"github.com/BergurHeimisson/md-viewer/internal/ansi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// listIndent is the number of spaces a nested list is indented by.
const listIndent = 2

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

// Renderer converts Markdown source to ANSI text.
//
// Width is the target column count for wrapped prose. Zero disables wrapping,
// which is what non-TTY output wants — piping to grep should preserve the
// document's own line structure.
type Renderer struct {
	Width int
}

// Render parses src as Markdown and returns ANSI-coloured terminal text.
func (r *Renderer) Render(src []byte) string {
	doc := md.Parser().Parse(text.NewReader(src))
	w := &walker{src: src}
	out := w.childBlocks(doc, r.Width)
	return strings.TrimLeft(out, "\n")
}

type walker struct {
	src []byte
}

// childBlocks renders every child of n as a block and joins them with a blank
// line. Each block ends in exactly one newline, so the join adds the gap.
func (w *walker) childBlocks(n ast.Node, width int) string {
	return w.joinBlocks(n, width, "\n")
}

// joinBlocks renders every child of n as a block, joined by sep. A tight list
// item passes an empty separator so its paragraph and any nested list sit on
// consecutive lines.
func (w *walker) joinBlocks(n ast.Node, width int, sep string) string {
	var parts []string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s := w.block(c, width); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, sep)
}

// block renders a single block-level node to text ending in one newline.
func (w *walker) block(n ast.Node, width int) string {
	switch n := n.(type) {
	case *ast.Heading:
		return w.heading(n, width)
	case *ast.Paragraph, *ast.TextBlock:
		return ansi.Wrap(w.inline(n), width) + "\n"
	case *ast.FencedCodeBlock:
		return w.codeBlock(n)
	case *ast.CodeBlock:
		return w.codeBlock(n)
	case *ast.Blockquote:
		return w.blockquote(n, width)
	case *ast.List:
		return w.list(n, width)
	case *ast.ThematicBreak:
		return ansi.BlackBright + strings.Repeat("─", ruleWidth(width)) + ansi.Reset + "\n"
	case *ast.HTMLBlock:
		return w.htmlBlock(n)
	case *extast.Table:
		return w.table(n)
	default:
		return w.childBlocks(n, width)
	}
}

func ruleWidth(width int) int {
	if width <= 0 || width > 60 {
		return 60
	}
	return width
}

func (w *walker) heading(n *ast.Heading, width int) string {
	var colour, marker string
	switch n.Level {
	case 1:
		colour, marker = ansi.CyanBold, "═══ "
	case 2:
		colour, marker = ansi.YellowBoldBright, "─── "
	default:
		colour, marker = ansi.PurpleBold, "▸ "
	}
	// The markers are multi-byte ("═══ " is 10 bytes, 4 cells), so width and
	// the hanging indent must both be measured in cells.
	pad := ansi.VisibleLen(marker)
	body := ansi.Wrap(w.inline(n), width-pad)
	return colour + marker + indentContinuation(body, pad) + ansi.Reset + "\n"
}

func (w *walker) codeBlock(n ast.Node) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		line := strings.TrimRight(string(seg.Value(w.src)), "\r\n")
		b.WriteString(ansi.Cyan + "  │ " + line + ansi.Reset + "\n")
	}
	if b.Len() == 0 {
		return ansi.Cyan + "  │ " + ansi.Reset + "\n"
	}
	return b.String()
}

func (w *walker) htmlBlock(n *ast.HTMLBlock) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(w.src))
	}
	if n.HasClosure() {
		b.Write(n.ClosureLine.Value(w.src))
	}
	out := strings.TrimRight(b.String(), "\n")
	if out == "" {
		return ""
	}
	return out + "\n"
}

func (w *walker) blockquote(n *ast.Blockquote, width int) string {
	const gutter = "▌ "
	body := w.childBlocks(n, width-2)
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		b.WriteString(ansi.GreenBold + gutter + ansi.Reset + line + "\n")
	}
	return b.String()
}

func (w *walker) list(n *ast.List, width int) string {
	var parts []string
	// Start is the parsed marker number for an ordered list, and CommonMark
	// allows it to be 0 — only a bullet list needs a default.
	number := n.Start
	if !n.IsOrdered() {
		number = 1
	}
	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		marker := bulletMarker()
		if n.IsOrdered() {
			marker = orderedMarker(number)
			number++
		} else if hasTaskCheckbox(item) {
			// The checkbox renders in place of the bullet, matching the old
			// Java renderer; a bullet as well reads as a double marker.
			marker = ""
		}
		parts = append(parts, w.listItem(item, marker, width, n.IsTight))
	}
	sep := ""
	if !n.IsTight {
		sep = "\n"
	}
	return strings.Join(parts, sep)
}

func bulletMarker() string {
	return ansi.Cyan + "•" + ansi.Reset + " "
}

func orderedMarker(n int) string {
	return strconv.Itoa(n) + ". "
}

// listItem renders one item: the marker on the first line, and every following
// line indented to line up beneath it.
func (w *walker) listItem(item ast.Node, marker string, width int, tight bool) string {
	pad := listIndent + ansi.VisibleLen(marker)
	sep := "\n"
	if tight {
		sep = ""
	}
	body := w.joinBlocks(item, width-pad, sep)
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return strings.Repeat(" ", listIndent) + marker + "\n"
	}
	lines := strings.Split(body, "\n")
	var b strings.Builder
	for i, line := range lines {
		if i == 0 {
			b.WriteString(strings.Repeat(" ", listIndent) + marker + line + "\n")
			continue
		}
		b.WriteString(strings.Repeat(" ", pad) + line + "\n")
	}
	return b.String()
}

func hasTaskCheckbox(item ast.Node) bool {
	first := item.FirstChild()
	if first == nil {
		return false
	}
	_, ok := first.FirstChild().(*extast.TaskCheckBox)
	return ok
}

// indentContinuation re-indents every line after the first by n spaces, so a
// wrapped heading lines up under its own text rather than its marker.
func indentContinuation(s string, n int) string {
	lines := strings.Split(s, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", n) + lines[i]
	}
	return strings.Join(lines, "\n")
}

// inline renders the inline children of n.
func (w *walker) inline(n ast.Node) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		b.WriteString(w.inlineNode(c))
	}
	return b.String()
}

func (w *walker) inlineNode(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Text:
		s := string(n.Segment.Value(w.src))
		switch {
		case n.HardLineBreak():
			return s + "\n"
		case n.SoftLineBreak():
			return s + " "
		}
		return s

	case *ast.String:
		return string(n.Value)

	case *ast.CodeSpan:
		return ansi.Cyan + w.rawText(n) + ansi.Reset

	case *ast.Emphasis:
		if n.Level >= 2 {
			return ansi.WhiteBold + w.inline(n) + ansi.Reset
		}
		return ansi.Yellow + w.inline(n) + ansi.Reset

	case *ast.Link:
		return w.inline(n) + ansi.Blue + " [" + string(n.Destination) + "]" + ansi.Reset

	case *ast.AutoLink:
		return ansi.Blue + string(n.URL(w.src)) + ansi.Reset

	case *ast.Image:
		return ansi.Purple + "🖼 " + ansi.Reset + w.inline(n) +
			ansi.Blue + " [" + string(n.Destination) + "]" + ansi.Reset

	case *ast.RawHTML:
		var b strings.Builder
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			b.Write(seg.Value(w.src))
		}
		return b.String()

	case *extast.Strikethrough:
		return ansi.BlackBright + w.inline(n) + ansi.Reset

	case *extast.TaskCheckBox:
		if n.IsChecked {
			return ansi.GreenBold + "☑" + ansi.Reset + " "
		}
		return ansi.BlackBright + "☐" + ansi.Reset + " "

	default:
		return w.inline(n)
	}
}

// rawText collects the literal source text under n, used for code spans where
// no child formatting applies.
func (w *walker) rawText(n ast.Node) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c := c.(type) {
		case *ast.Text:
			b.Write(c.Segment.Value(w.src))
		case *ast.String:
			b.Write(c.Value)
		default:
			b.WriteString(w.rawText(c))
		}
	}
	return b.String()
}
