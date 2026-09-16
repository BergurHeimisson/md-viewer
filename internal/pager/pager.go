// Package pager displays ANSI text one screen at a time on a terminal.
package pager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BergurHeimisson/md-viewer/internal/ansi"
	"golang.org/x/term"
)

// Page displays content one screen at a time, reading keys from in.
//
// It returns immediately after printing everything when the terminal is not
// interactive or the content already fits on one screen — the same shortcut
// `less -F` takes, which keeps short documents in the shell's scrollback.
func Page(out *os.File, in *os.File, content, title string) error {
	if !term.IsTerminal(int(out.Fd())) || !term.IsTerminal(int(in.Fd())) {
		_, err := io.WriteString(out, content)
		return err
	}

	cols, rows, err := term.GetSize(int(out.Fd()))
	if err != nil || rows < 3 {
		_, err := io.WriteString(out, content)
		return err
	}

	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) <= rows-1 {
		_, err := io.WriteString(out, content)
		return err
	}

	restore, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		_, err := io.WriteString(out, content)
		return err
	}

	s := &session{
		out:    out,
		reader: bufio.NewReader(in),
		lines:  lines,
		title:  title,
		cols:   cols,
		rows:   rows,
	}

	// Restore the terminal on every exit path, including a panic: leaving the
	// shell in raw mode with the alternate screen active makes it unusable.
	defer func() {
		fmt.Fprint(out, ansi.AutoWrapOn+ansi.CursorShow+ansi.AltScreenOff)
		_ = term.Restore(int(in.Fd()), restore)
	}()

	fmt.Fprint(out, ansi.AltScreenOn+ansi.AutoWrapOff+ansi.CursorHide)
	return s.run()
}

type session struct {
	out    io.Writer
	reader *bufio.Reader
	lines  []string
	title  string
	cols   int
	rows   int

	top    int // index of the first visible line
	search string
}

// view is the number of content lines visible, leaving one row for the status.
func (s *session) view() int { return s.rows - 1 }

// maxTop is the furthest the view can scroll: the last screenful.
func (s *session) maxTop() int {
	m := len(s.lines) - s.view()
	if m < 0 {
		return 0
	}
	return m
}

func (s *session) run() error {
	for {
		s.draw()
		key, err := s.readKey()
		if err != nil {
			return nil
		}
		if quit := s.handle(key); quit {
			return nil
		}
	}
}

// handle applies one key and reports whether the pager should exit.
func (s *session) handle(k key) bool {
	switch k {
	case keyQuit:
		return true
	case keyDown:
		s.scroll(1)
	case keyUp:
		s.scroll(-1)
	case keyPageDown:
		s.scroll(s.view())
	case keyPageUp:
		s.scroll(-s.view())
	case keyHalfDown:
		s.scroll(s.view() / 2)
	case keyHalfUp:
		s.scroll(-s.view() / 2)
	case keyTop:
		s.top = 0
	case keyBottom:
		s.top = s.maxTop()
	case keySearch:
		if q := s.prompt("/"); q != "" {
			s.search = q
			s.findNext(s.top + 1)
		}
	case keyNextMatch:
		s.findNext(s.top + 1)
	case keyPrevMatch:
		s.findPrev(s.top - 1)
	}
	return false
}

func (s *session) scroll(n int) {
	s.top += n
	if s.top < 0 {
		s.top = 0
	}
	if m := s.maxTop(); s.top > m {
		s.top = m
	}
}

func (s *session) findNext(from int) {
	if s.search == "" {
		return
	}
	for i := from; i < len(s.lines); i++ {
		if matches(s.lines[i], s.search) {
			s.top = min(i, s.maxTop())
			return
		}
	}
}

func (s *session) findPrev(from int) {
	if s.search == "" {
		return
	}
	for i := min(from, len(s.lines)-1); i >= 0; i-- {
		if matches(s.lines[i], s.search) {
			s.top = i
			return
		}
	}
}

// matches compares against the line's visible text so a search term is never
// derailed by the colour escapes wrapped around it.
func matches(line, needle string) bool {
	return strings.Contains(strings.ToLower(ansi.Strip(line)), strings.ToLower(needle))
}

func (s *session) draw() {
	var b strings.Builder
	b.WriteString(ansi.CursorHome + ansi.ClearScreen + ansi.CursorHome)
	for i := 0; i < s.view(); i++ {
		if idx := s.top + i; idx < len(s.lines) {
			b.WriteString(s.lines[idx])
		}
		b.WriteString(ansi.ClearToEndOfL + "\r\n")
	}
	b.WriteString(s.status())
	fmt.Fprint(s.out, b.String())
}

func (s *session) status() string {
	percent := 100
	if m := s.maxTop(); m > 0 {
		percent = s.top * 100 / m
	}
	left := fmt.Sprintf(" %s  %d%% ", s.title, percent)
	right := " SPACE next · b back · g/G · / search · q quit "

	gap := s.cols - ansi.VisibleLen(left) - ansi.VisibleLen(right)
	if gap < 1 {
		right = ""
		gap = max(s.cols-ansi.VisibleLen(left), 0)
	}
	return ansi.Reverse + left + strings.Repeat(" ", gap) + right + ansi.Reset
}

// prompt reads a line of input on the status row, returning "" if cancelled.
func (s *session) prompt(label string) string {
	var buf []rune
	for {
		fmt.Fprint(s.out, "\r"+ansi.ClearToEndOfL+label+string(buf))
		c, _, err := s.reader.ReadRune()
		if err != nil {
			return ""
		}
		switch c {
		case '\r', '\n':
			return string(buf)
		case 0x1b, 0x03: // Esc, Ctrl-C
			return ""
		case 0x7f, 0x08: // Backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
		default:
			if c >= 0x20 {
				buf = append(buf, c)
			}
		}
	}
}
