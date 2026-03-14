package com.mdviewer;

import com.formdev.flatlaf.FlatDarkLaf;
import com.formdev.flatlaf.FlatLaf;
import com.formdev.flatlaf.extras.FlatAnimatedLafChange;

import javax.swing.*;
import java.awt.*;
import java.io.File;
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
}
