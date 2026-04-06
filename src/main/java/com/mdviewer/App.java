package com.mdviewer;

import com.formdev.flatlaf.FlatDarkLaf;
import com.formdev.flatlaf.FlatLaf;
import com.formdev.flatlaf.extras.FlatAnimatedLafChange;

import javax.swing.*;
import java.awt.*;
import java.io.File;
import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.prefs.Preferences;

/**
 * Entry point for the mdViewer application.
 * Handles L&F setup, preference restoration, and command-line file argument.
 */
public class App {

    // Preference keys for persisting window state
    static final String PREF_X      = "window.x";
    static final String PREF_Y      = "window.y";
    static final String PREF_WIDTH  = "window.width";
    static final String PREF_HEIGHT = "window.height";
    static final String PREF_FONT      = "body.font";
    static final String PREF_FONT_SIZE = "body.font.size";

    /** Active colour scheme — loaded once at startup, readable app-wide. */
    static ColorScheme colors;

    public static void main(String[] args) {
        // Resolve file argument before touching the EDT
        final File initialFile;
        if (args.length > 0) {
            File candidate = new File(args[0]);
            initialFile = candidate.exists() ? candidate : null;
        } else {
            initialFile = null;
        }

        // Load colour scheme (user config file, or Dracula defaults)
        colors = ColorScheme.load();

        if (GraphicsEnvironment.isHeadless()) {
            runHeadless(initialFile);
            return;
        }

        // Use native file chooser whenever the platform supports it
        System.setProperty("flatlaf.useSystemFileChooser", "true");

        // Install FlatLaf Dracula theme before any component is created
        FlatLaf.registerCustomDefaultsSource("com.mdviewer");
        try {
            UIManager.setLookAndFeel(new FlatDraculaTheme());
        } catch (UnsupportedLookAndFeelException e) {
            // Fall back to plain dark L&F
            FlatDarkLaf.setup();
        }

        SwingUtilities.invokeLater(() -> {
            Preferences prefs = Preferences.userNodeForPackage(App.class);

            MainWindow window = new MainWindow(prefs);
            window.setVisible(true);

            if (initialFile != null) {
                window.openFile(initialFile);
            }
        });
    }

    private static void runHeadless(File file) {
        if (file == null) {
            System.err.println("Usage: mdviewer <file.md>");
            System.exit(1);
        }

        String markdown;
        try {
            markdown = Files.readString(file.toPath(), StandardCharsets.UTF_8);
        } catch (IOException e) {
            System.err.println("Cannot read file: " + file);
            System.exit(1);
            return;
        }

        String ansiOutput = new TerminalRenderer().render(markdown);

        try {
            Process pager = new ProcessBuilder("less", "-R")
                    .redirectInput(ProcessBuilder.Redirect.PIPE)
                    .inheritIO()
                    .redirectInput(ProcessBuilder.Redirect.PIPE)
                    .start();
            try (OutputStream stdin = pager.getOutputStream()) {
                stdin.write(ansiOutput.getBytes(StandardCharsets.UTF_8));
            }
            System.exit(pager.waitFor());
        } catch (IOException | InterruptedException e) {
            System.out.print(ansiOutput);
        }
    }
}
