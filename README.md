# md-viewer

A terminal Markdown viewer. Renders a `.md` file with ANSI colour and pages it
one screen at a time. Single Go binary, no runtime dependencies.

## Install

```sh
git clone https://github.com/BergurHeimisson/md-viewer.git
cd md-viewer
./install.sh
```

Builds with Go and installs the binary to `/usr/local/bin/md-viewer`. If a
previous Java install is present, `install.sh` removes it.

**Uninstall:**

```sh
sudo rm -f /usr/local/bin/md-viewer
```

## Usage

```sh
md-viewer path/to/file.md
```

The pager opens only when it is needed. A file that fits on one screen is
printed and left in your scrollback, and piping the output turns colour
wrapping and paging off entirely, so this works as expected:

```sh
md-viewer ARCHITECTURE.md | grep -i pager
```

### Flags

| Flag | Effect |
|---|---|
| `-headless` | Accepted and ignored — md-viewer is always headless |
| `-no-pager` | Print everything to stdout without paging |
| `-version` | Print the version and exit |

### Pager keys

| Key | Action |
|---|---|
| `SPACE`, `f`, `PgDn`, `Ctrl-F` | Next screen |
| `b`, `PgUp`, `Ctrl-B` | Previous screen |
| `d` / `u` | Half a screen down / up |
| `j` / `k`, `↓` / `↑` | One line |
| `g` / `G`, `Home` / `End` | Top / bottom |
| `/` | Search |
| `n` / `N` | Next / previous match |
| `q`, `Esc`, `Ctrl-C` | Quit |

## Rendering

GitHub-Flavored Markdown, via [goldmark](https://github.com/yuin/goldmark):
headings, emphasis, code spans and blocks, blockquotes, nested and ordered
lists, task lists, links, images, strikethrough, autolinked bare URLs, and
tables drawn as aligned box grids.

Prose wraps to your terminal width, capped at 100 columns. Code blocks and
tables are never re-flowed — they are pre-formatted, so the terminal clips them
horizontally instead.

## Development

```sh
go test ./...
go run ./cmd/md-viewer README.md
```

## Requirements

- Go 1.26+ to build
- macOS or Linux

## History

Versions before the Go rewrite were a Java Swing desktop app with a headless
terminal fallback. The GUI is gone; the terminal path became the whole tool.
See `ARCHITECTURE.md` for why.
