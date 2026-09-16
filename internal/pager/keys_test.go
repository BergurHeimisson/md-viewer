package pager

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// readKeyFrom decodes the first key from a literal byte sequence, the way it
// would arrive from a terminal in raw mode.
func readKeyFrom(t *testing.T, input string) key {
	t.Helper()
	s := &session{reader: bufio.NewReader(strings.NewReader(input))}
	k, err := s.readKey()
	if err != nil {
		t.Fatalf("readKey(%q) returned error: %v", input, err)
	}
	return k
}

func TestKeyDecoding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  key
	}{
		{"q quits", "q", keyQuit},
		{"Ctrl-C quits", "\x03", keyQuit},
		{"space pages down", " ", keyPageDown},
		{"f pages down", "f", keyPageDown},
		{"Ctrl-F pages down", "\x06", keyPageDown},
		{"b pages up", "b", keyPageUp},
		{"Ctrl-B pages up", "\x02", keyPageUp},
		{"d half down", "d", keyHalfDown},
		{"u half up", "u", keyHalfUp},
		{"j one line down", "j", keyDown},
		{"k one line up", "k", keyUp},
		{"return one line down", "\r", keyDown},
		{"g goes to the top", "g", keyTop},
		{"G goes to the bottom", "G", keyBottom},
		{"slash starts a search", "/", keySearch},
		{"n next match", "n", keyNextMatch},
		{"N previous match", "N", keyPrevMatch},
		{"unbound key does nothing", "z", keyNone},

		{"arrow up", "\x1b[A", keyUp},
		{"arrow down", "\x1b[B", keyDown},
		{"arrow right pages down", "\x1b[C", keyPageDown},
		{"arrow left pages up", "\x1b[D", keyPageUp},
		{"Home", "\x1b[H", keyTop},
		{"End", "\x1b[F", keyBottom},
		{"PgUp", "\x1b[5~", keyPageUp},
		{"PgDn", "\x1b[6~", keyPageDown},
		{"Home as CSI 1~", "\x1b[1~", keyTop},
		{"End as CSI 4~", "\x1b[4~", keyBottom},
		{"application-mode arrow up", "\x1bOA", keyUp},
		{"modified arrow up", "\x1b[1;5A", keyUp},

		// A lone Esc with nothing behind it is a quit, as in less.
		{"bare escape quits", "\x1b", keyQuit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readKeyFrom(t, tt.input); got != tt.want {
				t.Errorf("readKey(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPromptHandlesBackspace(t *testing.T) {
	s := &session{
		out:    io.Discard,
		reader: bufio.NewReader(strings.NewReader("deplo\x7foy\r")),
	}
	if got := s.prompt("/"); got != "deploy" {
		t.Errorf("prompt() = %q, want %q", got, "deploy")
	}
}

func TestPromptCancelsOnEscape(t *testing.T) {
	s := &session{out: io.Discard, reader: bufio.NewReader(strings.NewReader("abc\x1b"))}
	if got := s.prompt("/"); got != "" {
		t.Errorf("prompt() = %q, want empty after Esc", got)
	}
}
