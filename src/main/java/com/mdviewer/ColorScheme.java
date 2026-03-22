package com.mdviewer;

import java.awt.Color;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Properties;

/**
 * Holds the colour palette used for the HTML stylesheet and Swing component backgrounds.
 *
 * Defaults to the Dracula palette. Overrides are loaded from a user-editable
 * properties file whose location depends on the OS:
 *
 *   macOS   ~/Library/Application Support/mdviewer/colors.properties
 *   Linux   ~/.config/mdviewer/colors.properties
 *   Windows %APPDATA%\mdviewer\colors.properties
 *
 * Available keys (all optional — omitted keys fall back to Dracula defaults):
 *
 *   color.bg       page/editor background
 *   color.fg       body text
 *   color.surface  secondary background (code blocks, table header bg, row hover)
 *   color.muted    subdued text (blockquotes, placeholder subtitle)
 *   color.heading  h2–h6 color; also used for the drag-and-drop border highlight
 *   color.link     hyperlink color
 *   color.code     code / pre text color
 *   color.bold     **bold** text color
 *   color.italic   *italic* text color
 *   color.accent   table header text color
 *   color.del      ~~strikethrough~~ text color
 */
public class ColorScheme {

    // ── Dracula defaults ─────────────────────────────────────────────────────
    private static final String D_BG      = "#242424";
    private static final String D_FG      = "#E0E0E0";
    private static final String D_SURFACE = "#444444";
    private static final String D_MUTED   = "#888888";
    private static final String D_HEADING = "#B0B0B0";
    private static final String D_LINK    = "#B0B0B0";
    private static final String D_CODE    = "#E0E0E0";
    private static final String D_BOLD    = "#E0E0E0";
    private static final String D_ITALIC  = "#B0B0B0";
    private static final String D_ACCENT  = "#E0E0E0";
    private static final String D_DEL     = "#888888";

    // ── Public colour values ─────────────────────────────────────────────────
    public final String bg;
    public final String fg;
    public final String surface;
    public final String muted;
    public final String heading;
    public final String link;
    public final String code;
    public final String bold;
    public final String italic;
    public final String accent;
    public final String del;

    private ColorScheme(Properties p) {
        bg      = p.getProperty("color.bg",      D_BG);
        fg      = p.getProperty("color.fg",       D_FG);
        surface = p.getProperty("color.surface",  D_SURFACE);
        muted   = p.getProperty("color.muted",    D_MUTED);
        heading = p.getProperty("color.heading",  D_HEADING);
        link    = p.getProperty("color.link",     D_LINK);
        code    = p.getProperty("color.code",     D_CODE);
        bold    = p.getProperty("color.bold",     D_BOLD);
        italic  = p.getProperty("color.italic",   D_ITALIC);
        accent  = p.getProperty("color.accent",   D_ACCENT);
        del     = p.getProperty("color.del",      D_DEL);
    }

    /** Loads the scheme from the user config file, falling back to Dracula defaults. */
    public static ColorScheme load() {
        Properties p = new Properties();
        Path cfg = configFilePath();
        if (Files.exists(cfg)) {
            try (InputStream in = Files.newInputStream(cfg)) {
                p.load(in);
            } catch (IOException ignored) {}
        }
        return new ColorScheme(p);
    }

    /** Returns the platform-appropriate path for the colors.properties config file. */
    public static Path configFilePath() {
        String os = System.getProperty("os.name", "").toLowerCase();
        Path base;
        if (os.contains("mac")) {
            base = Path.of(System.getProperty("user.home"), "Library", "Application Support", "mdviewer");
        } else if (os.contains("win")) {
            String appData = System.getenv("APPDATA");
            base = appData != null
                    ? Path.of(appData, "mdviewer")
                    : Path.of(System.getProperty("user.home"), "mdviewer");
        } else {
            base = Path.of(System.getProperty("user.home"), ".config", "mdviewer");
        }
        return base.resolve("colors.properties");
    }

    // ── Convenience AWT conversions ──────────────────────────────────────────
    public Color awtBg()      { return Color.decode(bg); }
    public Color awtFg()      { return Color.decode(fg); }
    public Color awtHeading() { return Color.decode(heading); }
    public Color awtMuted()   { return Color.decode(muted); }
}
