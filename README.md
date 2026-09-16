# md-viewer

A terminal Markdown viewer. Renders a `.md` file with ANSI colour and pages it
one screen at a time. Single Go binary, no runtime dependencies.

## Install

```sh
git clone https://github.com/BergurHeimisson/md-viewer.git
cd md-viewer
./install.sh
```

Builds with Go and installs the binary to `/usr/local/bin/md-viewer`, which
needs your root password. If a previous Java install is present, `install.sh`
removes it.

**Without sudo:**

```sh
./install.sh --user
```

Installs to `go env GOBIN`, or `go env GOPATH`/bin if that is unset — typically
`~/go/bin`. Reach for this when sudo has no terminal to prompt on. The script
tells you if the directory is not on your `PATH`.

**Uninstall:**

```sh
./uninstall.sh
```

Finds every install this project could have created — system, `--user`, and
the old Java `mdviewer` — lists them, and asks before removing anything. Paths
you own are removed directly; only the root-owned ones ask for a password.
`--legacy` removes just the old Java install, `--yes` skips the prompt.

## Usage

```sh
md-viewer path/to/file.md
```

The pager opens only when it is needed. A file that fits on one screen is
printed and left in your scrollback, and piping the output turns colour,
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
| `q`, `Ctrl-C` | Quit |

## Rendering

GitHub-Flavored Markdown, via [goldmark](https://github.com/yuin/goldmark):
headings, emphasis, code spans and blocks, blockquotes, nested and ordered
lists, task lists, links, images, strikethrough, autolinked bare URLs, and
tables drawn as aligned box grids.

Prose wraps to your terminal width, capped at 100 columns. Code blocks and
tables are never re-flowed — they are pre-formatted, so the terminal clips them
horizontally instead.

Colour is emitted only when stdout is a terminal, and `NO_COLOR=1` turns it off
there too.

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
