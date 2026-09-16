# Architecture

## Overview

A single Go binary. No database, no network I/O, no config file. The pipeline
is: file bytes → goldmark AST → ANSI text → pager.

```
cmd/md-viewer/main.go        flags, file read, wire renderer → pager
internal/ansi/ansi.go        SGR constants, escape-aware VisibleLen and Wrap
internal/render/render.go    AST → ANSI (block and inline walkers)
internal/render/table.go     GFM table → box-drawn aligned grid
internal/pager/pager.go      raw-mode TTY pager, scrolling, search
internal/pager/keys.go       key and CSI escape decoding
```

## Key decisions

**Go instead of Java/Swing**
The Swing window was never used; the `-headless` path was. Delivering a
terminal pager through a Maven build, FlatLaf, an HTML renderer and a JVM
launcher cost startup latency and a Java 25 runtime dependency for no benefit.
A Go binary starts instantly and installs as one file.

**A hand-written AST walker, not goldmark's Renderer interface**
goldmark's renderer API is built around HTML-shaped streaming output. A
recursive walker maps far more directly onto the block concerns this tool has —
width, prefixes, indentation — and mirrors the Java visitor it replaced.

**Every block returns text ending in exactly one newline**
Containers join their children with one extra `\n`, which is what produces the
blank line between blocks. A container that adds a prefix (blockquote, list
item) renders its children against a reduced width and then prefixes each
resulting line. This single convention is what makes wrapping compose
correctly through arbitrary nesting; the Java renderer had no width concept and
could not nest at all.

**Tight lists join their item blocks with no separator**
`ast.List.IsTight` decides whether an item's paragraph and its nested sublist
are separated by a blank line. Without this, every nested bullet gained a
stray padded blank line above it.

**Escape-aware width measurement**
`ansi.VisibleLen` strips SGR sequences before counting, and counts CJK and
emoji ranges as two cells. Table columns, heading hanging indents and the pager
status line are all padded from this, so a bold cell or a `🖼` in a heading does
not skew the layout. Heading markers make this easy to get wrong: `"═══ "` is
10 bytes but 4 cells.

**Wrapping preserves runs of whitespace**
`ansi.Wrap` carries the gap before each word rather than rejoining with a
single space, so indentation and double spaces survive. Collapsing them would
make the same paragraph render differently at two terminal widths.

**Wrapping reopens the active colour after a break**
`ansi.Wrap` tracks the SGR code in effect where each word begins — not at the
read position, since the word's own escapes may already have changed it — and
on a break emits a reset before the newline and that code after it. Without
this, a wrapped emphasis bleeds into the pager's status line.

**Code blocks and tables are never wrapped**
They are pre-formatted; re-flowing destroys their meaning. The pager disables
terminal auto-wrap instead, so an over-wide table is clipped and stays aligned.

**A built-in pager rather than `less -R`**
`less` is absent on minimal Linux images, and its behaviour is shaped by the
user's `LESS` variable — if `-R` is missing, the escapes print literally. The
built-in pager also lets the bypass rules below be exact.

**The pager bypasses itself in three cases**
Output is printed straight to stdout when stdout or stdin is not a TTY, when
`--no-pager` is given, or when the content fits on one screen. The last case
mirrors `less -F`: a short file stays in the shell's scrollback after the
command returns. The first is what keeps `md-viewer x.md | grep foo` working.

**Alternate screen, auto-wrap off, and a deferred restore**
The pager switches to the alternate screen (`\e[?1049h`) so scrollback
survives, and restores raw mode, cursor, auto-wrap and the main screen from a
single `defer` covering every exit path including a panic. Leaving a terminal
in raw mode on the alternate screen makes the user's shell unusable.

**Piped output is plain, unwrapped text**
`Renderer.Width == 0` disables wrapping when stdout is not a TTY, and the
rendered text is run through `ansi.Strip` in the same case (and when `NO_COLOR`
is set). Both are needed for `md-viewer x.md | grep foo` to behave: wrapping
would break the document's own line structure, and escapes would make any
phrase straddling an emphasis boundary silently fail to match.

**A lone Esc does not quit**
Arrow and page keys send Esc as the first byte of a multi-byte sequence, and
nothing guarantees the terminal delivers the whole sequence in one `read(2)` —
over ssh or a slow tty the Esc can arrive alone. Treating it as quit would drop
the reader out of the document at random. Quit is `q` and `Ctrl-C`, which
cannot be split.

**The search cursor is tracked separately from the scroll position**
`top` is clamped to `maxTop`, so a match inside the last screenful does not
scroll to its own line. Resuming the next search from `top` would therefore
re-find that same match forever, making every later one unreachable. `match`
holds the real hit index and `n`/`N` advance from it.

**Each pager row is cleared before it is written, not after**
Auto-wrap is off, so writing a line at least as wide as the terminal parks the
cursor *in* the last column. A trailing `\e[K` there erases that cell — the
rightmost visible character of every over-wide row.

**Terminal size is re-read on each repaint**
A blocked `read` does not wake on `SIGWINCH`, so there is no signal handler; the
size is sampled in `draw` instead and a resize takes effect on the next
keypress. Already-rendered text keeps its original wrap width.

**`-headless` is a deliberate no-op**
The tool is always headless now, but the flag is in existing muscle memory and
scripts. Accepting and ignoring it is cheaper than breaking them.

## Dependencies

| Module | Purpose |
|---|---|
| `github.com/yuin/goldmark` | CommonMark parser; `extension.GFM` adds tables, strikethrough, task lists, autolinks |
| `golang.org/x/term` | Terminal size, TTY detection, raw mode |

## Testing

`go test ./...` covers three packages with table-driven tests:

- `internal/ansi` — `VisibleLen` over escapes and wide runes, `Wrap` at a fixed
  width, colour carried across a break, width 0 as a no-op.
- `internal/render` — one case per Markdown construct, table alignment and
  padding, nested list indentation, wrapping bounds.
- `internal/pager` — scroll arithmetic and clamping, search selection, status
  line width. Key handling is tested by feeding literal byte sequences
  (`"\x1b[6~"`) to the decoder, so no PTY is needed.

`gofmt` and `go vet` run in CI alongside the tests.

## Removed in the Go rewrite

The Swing window, FlatLaf theming, the `colors.properties` colour scheme (the
terminal path never read it — it used hardcoded ANSI constants), drag-and-drop,
arrow-key navigation between sibling `.md` files, window position persistence,
and the JVM launcher script.
