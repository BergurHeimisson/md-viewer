package com.mdviewer;

import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Properties;

import static org.junit.jupiter.api.Assertions.*;

class ColorSchemeTest {

    @Test
    void defaultsAreDraculaPalette() {
        ColorScheme cs = ColorScheme.load();
        assertEquals("#242424", cs.bg);
        assertEquals("#E0E0E0", cs.fg);
        assertEquals("#B0B0B0", cs.link);
    }

    @Test
    void configFilePathContainsMdviewer() {
        Path path = ColorScheme.configFilePath();
        assertTrue(path.toString().contains("mdviewer"),
                "Config path should contain 'mdviewer': " + path);
        assertEquals("colors.properties", path.getFileName().toString(),
                "Config file should be named colors.properties");
    }

    @Test
    void awtConversionsDoNotThrow() {
        ColorScheme cs = ColorScheme.load();
        assertNotNull(cs.awtBg());
        assertNotNull(cs.awtFg());
        assertNotNull(cs.awtHeading());
        assertNotNull(cs.awtMuted());
    }

    // ── configFilePath() OS branches ────────────────────────────────────────

    @Test
    void configFilePath_onLinux_isUnderDotConfig() {
        String savedOs = System.getProperty("os.name");
        try {
            System.setProperty("os.name", "Linux");
            Path path = ColorScheme.configFilePath();
            assertTrue(path.toString().contains(".config"),
                    "Linux path should contain .config: " + path);
            assertTrue(path.toString().contains("mdviewer"));
            assertEquals("colors.properties", path.getFileName().toString());
        } finally {
            System.setProperty("os.name", savedOs);
        }
    }

    @Test
    void configFilePath_onWindows_containsMdviewer() {
        String savedOs = System.getProperty("os.name");
        try {
            // APPDATA is not set on macOS/Linux → falls back to user.home/mdviewer
            System.setProperty("os.name", "Windows 10");
            Path path = ColorScheme.configFilePath();
            assertTrue(path.toString().contains("mdviewer"),
                    "Windows fallback path should contain mdviewer: " + path);
            assertEquals("colors.properties", path.getFileName().toString());
        } finally {
            System.setProperty("os.name", savedOs);
        }
    }

    // ── load() file-exists branch ────────────────────────────────────────────

    @Test
    void load_withCustomColorsFile_overridesBackground() throws IOException {
        Path configPath = ColorScheme.configFilePath();
        Files.createDirectories(configPath.getParent());

        Properties props = new Properties();
        props.setProperty("color.bg", "#112233");

        try (OutputStream out = Files.newOutputStream(configPath)) {
            props.store(out, null);
        }

        try {
            ColorScheme cs = ColorScheme.load();
            assertEquals("#112233", cs.bg);
            // Unset keys still fall back to defaults
            assertEquals("#E0E0E0", cs.fg);
            assertEquals("#B0B0B0", cs.heading);
        } finally {
            Files.deleteIfExists(configPath);
        }
    }

    @Test
    void allDefaultFieldsArePopulated() {
        ColorScheme cs = ColorScheme.load();
        assertNotNull(cs.bg);
        assertNotNull(cs.fg);
        assertNotNull(cs.surface);
        assertNotNull(cs.muted);
        assertNotNull(cs.heading);
        assertNotNull(cs.link);
        assertNotNull(cs.code);
        assertNotNull(cs.bold);
        assertNotNull(cs.italic);
        assertNotNull(cs.accent);
        assertNotNull(cs.del);
    }
}
