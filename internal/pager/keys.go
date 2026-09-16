package pager

import "bufio"

type key int

const (
	keyNone key = iota
	keyQuit
	keyUp
	keyDown
	keyPageUp
	keyPageDown
	keyHalfUp
	keyHalfDown
	keyTop
	keyBottom
	keySearch
	keyNextMatch
	keyPrevMatch
)

func (s *session) readKey() (key, error) {
	c, err := s.reader.ReadByte()
	if err != nil {
		return keyNone, err
	}
	if c == 0x1b {
		return s.readEscape(), nil
	}
	return decode(c), nil
}

func decode(c byte) key {
	switch c {
	case 'q', 'Q', 0x03: // Ctrl-C
		return keyQuit
	case ' ', 'f', 0x06: // Ctrl-F
		return keyPageDown
	case 'b', 0x02: // Ctrl-B
		return keyPageUp
	case 'd', 0x04: // Ctrl-D
		return keyHalfDown
	case 'u', 0x15: // Ctrl-U
		return keyHalfUp
	case 'j', '\r', '\n':
		return keyDown
	case 'k':
		return keyUp
	case 'g':
		return keyTop
	case 'G':
		return keyBottom
	case '/':
		return keySearch
	case 'n':
		return keyNextMatch
	case 'N':
		return keyPrevMatch
	}
	return keyNone
}

// readEscape decodes a CSI sequence after the leading Esc has been consumed.
// A bare Esc — nothing buffered behind it — quits, matching less.
func (s *session) readEscape() key {
	if s.reader.Buffered() == 0 {
		return keyQuit
	}
	c, err := s.reader.ReadByte()
	if err != nil {
		return keyQuit
	}
	if c != '[' && c != 'O' {
		return keyNone
	}
	return decodeCSI(s.reader)
}

// decodeCSI reads the parameter and final bytes of a CSI sequence.
func decodeCSI(r *bufio.Reader) key {
	var params []byte
	for {
		c, err := r.ReadByte()
		if err != nil {
			return keyNone
		}
		if c >= '0' && c <= '9' || c == ';' {
			params = append(params, c)
			continue
		}
		return csiKey(string(params), c)
	}
}

func csiKey(params string, final byte) key {
	switch final {
	case 'A':
		return keyUp
	case 'B':
		return keyDown
	case 'C':
		return keyPageDown
	case 'D':
		return keyPageUp
	case 'H':
		return keyTop
	case 'F':
		return keyBottom
	case '~':
		switch params {
		case "5":
			return keyPageUp
		case "6":
			return keyPageDown
		case "1", "7":
			return keyTop
		case "4", "8":
			return keyBottom
		}
	}
	return keyNone
}
