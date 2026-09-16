# md-viewer — Go rewrite design

**Date:** 2026-09-16
**Status:** approved

## Goal

Replace the Java/Swing `mdViewer` with a single Go binary named `md-viewer`
that renders a Markdown file to ANSI and pages it one screen at a time. The
GUI is dropped entirely; the headless terminal path becomes the whole product.

## Why

The Swing window was never used. The `-headless` path — CommonMark → ANSI,
piped into `less -R` — is the only mode in daily use. Carrying a Maven build,
FlatLaf, an HTML renderer and a JVM launcher to deliver a terminal pager is
pure cost: slow startup, a Java 25 runtime dependency, and 1,600 lines of code
where ~700 will do.

## Scope

**In:** ANSI Markdown rendering, a built-in pager, a Go build, install script,
docs, CI. Deletion of all Java sources, `pom.xml`, and the JVM launcher.

**Out:** the Swing GUI, FlatLaf theming, `colors.properties`, drag-and-drop,
sibling-file arrow navigation, window position persistence, the markitdown
integration spec (untouched, still a future feature).

## Architecture

```
cmd/md-viewer/main.go        flag parsing, file read, wire renderer → pager
internal/ansi/ansi.go        SGR constants, escape-aware VisibleLen and Wrap
internal/render/render.go    goldmark AST → ANSI (block + inline walkers)
internal/render/table.go     GFM table → box-drawn aligned grid
internal/pager/pager.go      raw-mode TTY pager, TTY detection
```

Data flow: `os.ReadFile` → `goldmark.Parse` → `render.Renderer.Render` →
`pager.Page`. No network, no config file, no persistent state.

### Parsing

`github.com/yuin/goldmark` with `extension.GFM`, which supplies tables,
strikethrough, linkified bare URLs and task-list items — matching the
CommonMark extension set the Java `TerminalRenderer` used. Heading anchors are
dropped; they were meaningless in a terminal.

Rendering does not use goldmark's own `Renderer` interface. A hand-written
recursive walker over the AST mirrors the structure of the Java visitor and
makes block-level concerns (wrapping, prefixing, indentation) straightforward.

### Rendering contract

Every block renders to a string ending in exactly one `\n`. Containers join
their child blocks with one extra `\n`, producing a blank line between blocks.
A block is rendered against a target width; containers that add a prefix
(blockquotes, list items) pass a reduced width to their children and prefix
each resulting line. This is what makes wrapping compose correctly through
nesting — the Java renderer had no width concept at all.

Visual style is a faithful port of `TerminalRenderer`:

| Element | Rendering |
|---|---|
| H1 | cyan bold, `═══ ` prefix |
| H2 | bright yellow bold, `─── ` prefix |
| H3+ | purple bold, `▸ ` prefix |
| Code block | cyan, `  │ ` gutter per line |
| Blockquote | green bold `▌ ` gutter per line |
| Bullet list | cyan `•`, two-space indent per nesting level |
| Ordered list | `N.`, honouring the start number |
| Task item | green `☑` / bright-black `☐` in place of the bullet |
| Link | text followed by blue ` [destination]` |
| Image | purple `🖼 `, alt text, blue ` [destination]` |
| Strikethrough | bright black |
| Thematic break | bright black `─` to the wrap width |
| Table | bright-black `┌┬┐` grid, bold header, columns padded to content |

Inline HTML and HTML blocks are emitted literally, as the Java version did,
so `<min>`/`<max>` in prose survive.

### Wrapping

Paragraphs, headings and blockquote text wrap to the terminal width. Code
blocks and tables never wrap — they are pre-formatted, and re-flowing them
destroys meaning; they are clipped horizontally by the terminal instead.

Wrapping is escape-aware: `ansi.VisibleLen` ignores SGR sequences when
measuring, and `ansi.Wrap` tracks the active SGR code so that breaking a line
mid-emphasis emits a reset before the newline and re-applies the code after
it. Without this, colour bleeds into the pager's status line.

Width comes from `term.GetSize` on stdout, capped at 100 columns so prose
stays readable on a wide terminal. When stdout is not a TTY, width is zero and
no wrapping happens — `md-viewer x.md | grep foo` must see the original lines.

### Pager

Built in, with no dependency on `less`. Justification: `less` is absent on
minimal Linux images, and its behaviour is tuned by the user's `LESS`
environment variable, which has bitten this tool before (`-R` must be present
or the escapes print literally).

- Alternate screen (`\e[?1049h`) so the shell scrollback is restored on exit.
- Auto-wrap disabled (`\e[?7l`) so an over-wide table clips rather than
  corrupting the layout.
- Raw mode via `golang.org/x/term`, restored by `defer` on every exit path
  including panic.

Keys: `Space`/`f`/`PgDn`/`Ctrl-F` next screen, `b`/`PgUp`/`Ctrl-B` previous,
`j`/`↓` and `k`/`↑` by line, `g`/`Home` top, `G`/`End` bottom, `/` search,
`n`/`N` next/previous match, `q`/`Esc`/`Ctrl-C` quit.

Search matches against the visible text of a line (escapes stripped),
case-insensitively, and jumps to the matching line.

The status line shows the filename and scroll percentage in reverse video.

**Bypass rules.** The pager is skipped — content printed straight to stdout —
when stdout is not a TTY, when `--no-pager` is given, or when the content
fits on one screen. The last case matches `less -F` and keeps short files
visible in scrollback after the command returns.

### CLI

```
md-viewer [flags] <file.md>

  -headless     accepted and ignored (the tool is always headless)
  -no-pager     print everything to stdout without paging
  -version      print version and exit
```

`-headless` is a deliberate no-op rather than an error: it is the invocation
in the user's muscle memory and in existing scripts.

Errors — missing argument, unreadable file — go to stderr with exit status 1.

## Testing

Go table-driven tests, ported from `TerminalRendererTest` and extended:

- `internal/render`: one case per Markdown construct, asserting on the ANSI
  output; table alignment with multi-byte and emphasised cell content; nested
  list indentation; wrapping at a fixed width.
- `internal/ansi`: `VisibleLen` against strings containing SGR sequences;
  `Wrap` preserving active colour across a break; `Wrap` with width 0.
- `internal/pager`: page arithmetic (`next`, `prev`, clamping at both ends),
  fits-on-one-screen detection, search hit selection. Key decoding is tested
  against byte sequences, so no PTY is needed.

`go vet ./...` and `gofmt -l` run in CI alongside `go test ./...`.

## Migration

Deleted: `pom.xml`, `src/`, `target/`, the `mdviewer` shell launcher,
`src/main/resources/com/mdviewer/FlatDraculaTheme.properties`.

Rewritten: `install.sh` (build with `go build`, install the binary to
`/usr/local/bin/md-viewer`, no lib directory and no launcher script needed),
`README.md`, `ARCHITECTURE.md`, `.gitignore`, and both GitHub workflows
(`setup-go` in place of `setup-java`; the release workflow cross-compiles
darwin/linux × amd64/arm64 and attaches the binaries).

`features/headless/SPEC.md` is superseded and removed;
`features/markidown_integration/SPEC.md` stays.

The old `mdviewer` command name is not preserved. The user asked for the
artifact to be renamed to `md-viewer`; a lingering alias would defeat that.
Uninstalling the Java version is `sudo rm -rf /usr/local/lib/mdviewer
/usr/local/bin/mdviewer`, called out in the README.
