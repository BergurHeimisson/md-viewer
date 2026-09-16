// Command md-viewer renders a Markdown file to ANSI and pages it one screen
// at a time.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BergurHeimisson/md-viewer/internal/ansi"
	"github.com/BergurHeimisson/md-viewer/internal/pager"
	"github.com/BergurHeimisson/md-viewer/internal/render"
	"golang.org/x/term"
)

// Version is overridden at build time with -ldflags "-X main.Version=...".
var Version = "dev"

// maxWidth caps prose wrapping so long lines stay readable on a wide terminal.
const maxWidth = 100

func main() {
	// -headless is accepted and ignored: md-viewer is always headless now, but
	// the flag is in existing muscle memory and scripts.
	flag.Bool("headless", false, "accepted and ignored (md-viewer is always headless)")
	noPager := flag.Bool("no-pager", false, "print everything to stdout without paging")
	showVersion := flag.Bool("version", false, "print version and exit")

	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println("md-viewer", Version)
		return
	}

	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}

	path := flag.Arg(0)
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "md-viewer: %v\n", err)
		os.Exit(1)
	}

	r := &render.Renderer{Width: wrapWidth()}
	content := r.Render(src)
	if !colourEnabled() {
		content = ansi.Strip(content)
	}

	if *noPager {
		fmt.Print(content)
		return
	}

	if err := pager.Page(os.Stdout, os.Stdin, content, filepath.Base(path)); err != nil {
		fmt.Fprintf(os.Stderr, "md-viewer: %v\n", err)
		os.Exit(1)
	}
}

// colourEnabled reports whether to emit ANSI escapes. Piped output must stay
// greppable, and NO_COLOR is honoured for terminals that want plain text.
// See https://no-color.org.
func colourEnabled() bool {
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// wrapWidth is the terminal width, capped, or zero when stdout is not a
// terminal — piped output keeps the document's own line structure.
func wrapWidth() int {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return 0
	}
	cols, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || cols <= 0 {
		return 0
	}
	return min(cols, maxWidth)
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: md-viewer [flags] <file.md>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Flags:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Pager keys:")
	fmt.Fprintln(os.Stderr, "  SPACE, f, PgDn, Ctrl-F   next screen")
	fmt.Fprintln(os.Stderr, "  b, PgUp, Ctrl-B          previous screen")
	fmt.Fprintln(os.Stderr, "  d / u                    half screen down / up")
	fmt.Fprintln(os.Stderr, "  j / k, arrow down / up   one line")
	fmt.Fprintln(os.Stderr, "  g / G, Home / End        top / bottom")
	fmt.Fprintln(os.Stderr, "  / then n / N             search, next / previous match")
	fmt.Fprintln(os.Stderr, "  q, Ctrl-C                quit")
}
