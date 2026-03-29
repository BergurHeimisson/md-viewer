# mdViewer

A minimalistic Java Swing application for viewing Markdown files with a Dracula-themed dark UI.

## Features

- Drag-and-drop a `.md` or `.markdown` file onto the window to open it
- Pass a file path as a command-line argument
- Cycle through all `.md` files in the same folder with the arrow keys
- Scroll content with up/down arrow keys
- Open files via `Cmd+O`
- Rendered Markdown with support for tables, task lists, strikethrough, autolinks, and heading anchors
- Dracula colour theme throughout — FlatLaf for the chrome, inline CSS for the content
- Fully configurable colour scheme via a plain-text properties file
- Window position and size remembered across restarts

## Requirements

- Java 17+
- Maven 3.6+ (to build)
- macOS or Linux (Windows is not supported)

## Installation

Clone the repo and run the install script:

```sh
git clone https://github.com/BergurHeimisson/md-viewer.git
cd md-viewer
./install.sh
```

Installs to `/usr/local/lib/mdviewer/` and symlinks the launcher into `/usr/local/bin/` — no PATH or shell configuration needed after that.

**Uninstall:**

```sh
sudo rm -rf /usr/local/lib/mdviewer /usr/local/bin/mdviewer
```

## Usage

```sh
mdviewer path/to/file.md
```

Or launch with no argument and drop a file onto the window.

## Colour scheme

All colours can be overridden via a properties file. Create it at:

- **macOS:** `~/Library/Application Support/mdviewer/colors.properties`
- **Linux:** `~/.config/mdviewer/colors.properties`

Any omitted keys fall back to the Dracula defaults. Available keys:

```properties
color.bg=#282a36       # page/editor background
color.fg=#f8f8f2       # body text
color.surface=#44475a  # code block background, table row hover, borders
color.muted=#6272a4    # blockquotes, subdued text
color.heading=#bd93f9  # h2–h6 headings
color.link=#8be9fd     # hyperlinks
color.code=#50fa7b     # code / pre text
color.bold=#ffb86c     # **bold** text
color.italic=#f1fa8c   # *italic* text
color.accent=#ff79c6   # table header text
color.del=#ff5555      # ~~strikethrough~~ text
```

Restart the app after editing to apply changes.

## Development

Build and run without installing:

```sh
mvn package
./mdviewer path/to/file.md
```

## Key bindings

| Key         | Action                              |
|-------------|-------------------------------------|
| `↑` / `↓`  | Scroll up / down                    |
| `←` / `→`  | Previous / next file in same folder |
| `Cmd+O`     | Open file via dialog                |
| `Cmd+R`     | Reload current file                 |
| `Cmd+W`     | Quit                                |
| `Escape`    | Quit                                |
