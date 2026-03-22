package com.mdviewer;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.IOException;
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
        assertTrue(path.getFileName().toString().equals("colors.properties"),
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
}
