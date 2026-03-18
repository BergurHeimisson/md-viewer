package com.mdviewer;

import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class MarkdownRendererTest {

    @BeforeAll
    static void initColors() {
        App.colors = ColorScheme.load();
    }

    @Test
    void renderProducesHtmlDocument() {
        String html = new MarkdownRenderer().render("# Hello");
        assertTrue(html.contains("<!DOCTYPE html>"));
        assertTrue(html.contains("<h1"));
        assertTrue(html.contains("Hello"));
    }

    @Test
    void boldAndItalicAreRendered() {
        String html = new MarkdownRenderer().render("**bold** and *italic*");
        assertTrue(html.contains("<strong>"));
        assertTrue(html.contains("<em>"));
    }

    @Test
    void codeBlockIsRendered() {
        String html = new MarkdownRenderer().render("```\nSystem.out.println();\n```");
        assertTrue(html.contains("<pre>") || html.contains("<code>"));
    }

    @Test
    void emptyInputProducesValidHtml() {
        String html = new MarkdownRenderer().render("");
        assertTrue(html.contains("<html>"));
        assertTrue(html.contains("</html>"));
    }

    @Test
    void fontSettingsAreReflectedInCss() {
        MarkdownRenderer r = new MarkdownRenderer();
        r.setFontFamily("monospace");
        r.setFontSize(16);
        String html = r.render("test");
        assertTrue(html.contains("monospace"));
        assertTrue(html.contains("16px"));
    }
}
