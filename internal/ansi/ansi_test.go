package ansi

import "testing"

func TestStripRemovesEscapes(t *testing.T) {
	in := CyanBold + "hello " + Reset + Yellow + "world" + Reset
	if got, want := Strip(in), "hello world"; got != want {
		t.Errorf("Strip() = %q, want %q", got, want)
	}
}

func TestVisibleLen(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"plain", "hello", 5},
		{"coloured", Cyan + "hello" + Reset, 5},
		{"empty", "", 0},
		{"escapes only", Reset + Cyan, 0},
		{"box drawing is one cell", "┌──┐", 4},
		{"emoji is two cells", "🖼", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VisibleLen(tt.in); got != tt.want {
				t.Errorf("VisibleLen(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestWrapZeroWidthIsUnchanged(t *testing.T) {
	in := "a very long line that would otherwise be broken into several pieces"
	if got := Wrap(in, 0); got != in {
		t.Errorf("Wrap(_, 0) = %q, want input unchanged", got)
	}
}

func TestWrapBreaksOnSpaces(t *testing.T) {
	got := Wrap("one two three four five", 9)
	want := "one two\nthree\nfour five"
	if got != want {
		t.Errorf("Wrap() = %q, want %q", got, want)
	}
}

func TestWrapKeepsEveryLineWithinWidth(t *testing.T) {
	const width = 20
	in := "the quick brown fox jumps over the lazy dog and keeps on running"
	for _, line := range splitLines(Wrap(in, width)) {
		if n := VisibleLen(line); n > width {
			t.Errorf("line %q is %d cells, want <= %d", line, n, width)
		}
	}
}

// An over-long word has nowhere to break, so it is allowed to overflow rather
// than being cut in half.
func TestWrapDoesNotSplitAnOverlongWord(t *testing.T) {
	in := "short supercalifragilisticexpialidocious"
	got := Wrap(in, 10)
	want := "short\nsupercalifragilisticexpialidocious"
	if got != want {
		t.Errorf("Wrap() = %q, want %q", got, want)
	}
}

// Colour must not bleed past a wrap point: the break closes the active code
// and reopens it on the next line.
func TestWrapCarriesColourAcrossBreak(t *testing.T) {
	in := Yellow + "alpha bravo charlie" + Reset
	got := Wrap(in, 11)
	want := Yellow + "alpha bravo" + Reset + "\n" + Yellow + "charlie" + Reset
	if got != want {
		t.Errorf("Wrap() = %q, want %q", got, want)
	}
}

func TestWrapPreservesExistingNewlines(t *testing.T) {
	got := Wrap("one two\nthree four", 80)
	if want := "one two\nthree four"; got != want {
		t.Errorf("Wrap() = %q, want %q", got, want)
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// Whitespace runs must survive a wrap, so the same paragraph does not render
// differently at two terminal widths.
func TestWrapPreservesSpaceRuns(t *testing.T) {
	in := "alpha  beta   gamma delta epsilon"
	if got := Wrap(in, 100); got != in {
		t.Errorf("unwrapped: got %q, want %q", got, in)
	}
	got := Wrap(in, 20)
	want := "alpha  beta   gamma\ndelta epsilon"
	if got != want {
		t.Errorf("wrapped: got %q, want %q", got, want)
	}
}

func TestWrapKeepsLeadingIndent(t *testing.T) {
	got := Wrap("    indented text that will need breaking", 20)
	want := "    indented text\nthat will need\nbreaking"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
