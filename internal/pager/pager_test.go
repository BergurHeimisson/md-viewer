package pager

import (
	"fmt"
	"strings"
	"testing"

	"github.com/BergurHeimisson/md-viewer/internal/ansi"
)

// newSession builds a session over n numbered lines with a 10-row terminal,
// so the view is 9 lines and maxTop is n-9.
func newSession(n int) *session {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i)
	}
	return &session{lines: lines, cols: 80, rows: 10, title: "test.md"}
}

func TestViewLeavesARowForTheStatusLine(t *testing.T) {
	s := newSession(50)
	if got, want := s.view(), 9; got != want {
		t.Errorf("view() = %d, want %d", got, want)
	}
}

func TestMaxTopIsTheLastScreenful(t *testing.T) {
	tests := []struct {
		lines int
		want  int
	}{
		{50, 41},
		{10, 1},
		{9, 0},
		{3, 0}, // shorter than one screen: cannot scroll at all
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.lines), func(t *testing.T) {
			if got := newSession(tt.lines).maxTop(); got != tt.want {
				t.Errorf("maxTop() with %d lines = %d, want %d", tt.lines, got, tt.want)
			}
		})
	}
}

func TestScrollClampsAtBothEnds(t *testing.T) {
	s := newSession(50)

	s.scroll(-5)
	if s.top != 0 {
		t.Errorf("scrolling up from the top moved to %d, want 0", s.top)
	}

	s.scroll(1000)
	if s.top != s.maxTop() {
		t.Errorf("scrolling past the end moved to %d, want %d", s.top, s.maxTop())
	}
}

func TestPagingKeys(t *testing.T) {
	tests := []struct {
		name    string
		keys    []key
		wantTop int
	}{
		{"page down", []key{keyPageDown}, 9},
		{"page down twice", []key{keyPageDown, keyPageDown}, 18},
		{"page down then up", []key{keyPageDown, keyPageUp}, 0},
		{"one line", []key{keyDown, keyDown}, 2},
		{"half screen", []key{keyHalfDown}, 4},
		{"bottom", []key{keyBottom}, 41},
		{"bottom then top", []key{keyBottom, keyTop}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSession(50)
			for _, k := range tt.keys {
				if quit := s.handle(k); quit {
					t.Fatalf("key %v quit unexpectedly", k)
				}
			}
			if s.top != tt.wantTop {
				t.Errorf("top = %d, want %d", s.top, tt.wantTop)
			}
		})
	}
}

func TestQuitKeyStops(t *testing.T) {
	if !newSession(50).handle(keyQuit) {
		t.Error("keyQuit should stop the pager")
	}
}

func TestSearchJumpsToTheMatch(t *testing.T) {
	s := newSession(50)
	s.search = "line 23"
	s.findNext(0)
	if s.top != 23 {
		t.Errorf("top = %d, want 23", s.top)
	}
}

func TestSearchIsCaseInsensitiveAndIgnoresColour(t *testing.T) {
	s := newSession(5)
	s.lines[3] = ansi.CyanBold + "Deployment" + ansi.Reset + " notes"
	s.rows = 3 // view of 2, so a jump to line 3 is possible
	s.search = "deployment"
	s.findNext(0)
	if s.top != 3 {
		t.Errorf("top = %d, want 3", s.top)
	}
}

func TestSearchWithNoMatchLeavesThePositionAlone(t *testing.T) {
	s := newSession(50)
	s.top = 12
	s.search = "nothing here"
	s.findNext(0)
	if s.top != 12 {
		t.Errorf("top = %d, want it unchanged at 12", s.top)
	}
}

func TestFindPrevSearchesBackwards(t *testing.T) {
	s := newSession(50)
	s.top = 30
	s.search = "line 2"
	s.findPrev(s.top - 1)
	if s.top != 29 { // "line 29" contains "line 2"
		t.Errorf("top = %d, want 29", s.top)
	}
}

func TestStatusLineFitsTheTerminalWidth(t *testing.T) {
	for _, cols := range []int{20, 40, 80, 200} {
		s := newSession(50)
		s.cols = cols
		if got := ansi.VisibleLen(s.status()); got != cols {
			t.Errorf("status line at %d cols is %d cells, want %d", cols, got, cols)
		}
	}
}

func TestStatusLineShowsTitleAndPercent(t *testing.T) {
	s := newSession(50)
	s.top = s.maxTop()
	got := ansi.Strip(s.status())
	if !strings.Contains(got, "test.md") {
		t.Errorf("status %q should name the file", got)
	}
	if !strings.Contains(got, "100%") {
		t.Errorf("status %q should read 100%% at the bottom", got)
	}
}
