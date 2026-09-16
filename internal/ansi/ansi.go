// Package ansi provides SGR colour constants and escape-aware text measurement.
package ansi

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	Reset = "\033[0m"

	Green  = "\033[0;32m"
	Yellow = "\033[0;33m"
	Blue   = "\033[0;34m"
	Purple = "\033[0;35m"
	Cyan   = "\033[0;36m"

	GreenBold  = "\033[1;32m"
	PurpleBold = "\033[1;35m"
	CyanBold   = "\033[1;36m"
	WhiteBold  = "\033[1;37m"

	BlackBright = "\033[0;90m"

	YellowBoldBright = "\033[1;93m"

	Reverse = "\033[7m"
)

// Terminal control sequences used by the pager.
const (
	AltScreenOn   = "\033[?1049h"
	AltScreenOff  = "\033[?1049l"
	AutoWrapOff   = "\033[?7l"
	AutoWrapOn    = "\033[?7h"
	CursorHide    = "\033[?25l"
	CursorShow    = "\033[?25h"
	ClearScreen   = "\033[2J"
	CursorHome    = "\033[H"
	ClearToEndOfL = "\033[K"
)

// Strip removes every SGR escape sequence from s.
func Strip(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '\033' {
			if j := escapeEnd(s, i); j > i {
				i = j
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// VisibleLen counts the printable width of s in cells, ignoring escape
// sequences. Runes outside the Basic Multilingual Plane and East Asian wide
// ranges are treated as one cell, which is accurate for the box-drawing and
// symbol characters this renderer emits.
func VisibleLen(s string) int {
	n := 0
	for _, r := range Strip(s) {
		n += runeWidth(r)
	}
	return n
}

func runeWidth(r rune) int {
	switch {
	case r == '\n' || r == '\r':
		return 0
	case unicode.Is(unicode.Mn, r):
		return 0
	case isWide(r):
		return 2
	default:
		return 1
	}
}

// isWide reports whether r occupies two terminal cells. The ranges cover CJK,
// Hangul and the emoji blocks that show up in Markdown prose.
func isWide(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo
		r >= 0x2E80 && r <= 0xA4CF, // CJK radicals through Yi
		r >= 0xAC00 && r <= 0xD7A3, // Hangul syllables
		r >= 0xF900 && r <= 0xFAFF, // CJK compatibility ideographs
		r >= 0xFE30 && r <= 0xFE6F, // CJK compatibility forms
		r >= 0xFF00 && r <= 0xFF60, // Fullwidth forms
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1F64F, // Misc symbols and pictographs, emoticons
		r >= 0x1F900 && r <= 0x1F9FF, // Supplemental symbols and pictographs
		r >= 0x20000 && r <= 0x3FFFD: // CJK extension B onwards
		return true
	}
	return false
}

// escapeEnd returns the index just past the escape sequence starting at i, or
// i if the bytes at i are not a recognised sequence.
func escapeEnd(s string, i int) int {
	if i+1 >= len(s) || s[i+1] != '[' {
		return i
	}
	for j := i + 2; j < len(s); j++ {
		if c := s[j]; c >= '@' && c <= '~' {
			return j + 1
		}
	}
	return i
}

// Wrap breaks s to at most width visible cells per line, splitting on spaces.
// A width of zero returns s unchanged.
//
// Colour is carried across a break: the active SGR code is closed with a reset
// before the newline and re-opened after it, so a wrapped emphasis does not
// bleed into whatever the caller prints next.
func Wrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, wrapLine(line, width))
	}
	return strings.Join(out, "\n")
}

func wrapLine(line string, width int) string {
	if VisibleLen(line) <= width {
		return line
	}

	var (
		b       strings.Builder
		cur     int    // visible cells written on the current output line
		active  string // SGR code in effect at the read position
		word    strings.Builder
		wordLen int
		// SGR code in effect where the pending word begins. A break happens
		// before the word, so this — not active, which the word's own escapes
		// may already have changed — is what must be closed and reopened.
		wordStart string
	)

	// appendToWord records the colour state at the start of each new word.
	appendToWord := func(s string) {
		if word.Len() == 0 {
			wordStart = active
		}
		word.WriteString(s)
	}

	flushWord := func() {
		if word.Len() == 0 {
			return
		}
		// A word that cannot fit on any line is placed as-is and allowed to
		// overflow; breaking inside it would be worse than a ragged edge.
		if cur > 0 && cur+1+wordLen > width {
			if wordStart != "" {
				b.WriteString(Reset)
			}
			b.WriteByte('\n')
			b.WriteString(wordStart)
			cur = 0
		} else if cur > 0 && wordLen > 0 {
			b.WriteByte(' ')
			cur++
		}
		b.WriteString(word.String())
		cur += wordLen
		word.Reset()
		wordLen = 0
	}

	for i := 0; i < len(line); {
		if line[i] == '\033' {
			if j := escapeEnd(line, i); j > i {
				code := line[i:j]
				// Append before updating active, so a word that opens with an
				// escape records the state preceding it and the reopened code
				// is not emitted twice.
				appendToWord(code)
				if code == Reset {
					active = ""
				} else {
					active = code
				}
				i = j
				continue
			}
		}
		if line[i] == ' ' {
			flushWord()
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(line[i:])
		appendToWord(line[i : i+size])
		wordLen += runeWidth(r)
		i += size
	}
	flushWord()
	return b.String()
}
