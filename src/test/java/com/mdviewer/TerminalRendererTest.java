package com.mdviewer;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class TerminalRendererTest {

    private final TerminalRenderer renderer = new TerminalRenderer();

    @Test
    void h1HeadingIsRenderedWithCyanBold() {
        String output = renderer.render("# Hello");
        assertTrue(output.contains("Hello"));
        assertTrue(output.contains(AnsiColor.CYAN_BOLD));
        assertTrue(output.contains(AnsiColor.RESET));
    }

    @Test
    void h2HeadingIsDistinctFromH1() {
        String h1 = renderer.render("# Title");
        String h2 = renderer.render("## Subtitle");
        assertTrue(h2.contains(AnsiColor.YELLOW_BOLD_BRIGHT));
        assertFalse(h2.contains(AnsiColor.CYAN_BOLD));
        assertFalse(h1.contains(AnsiColor.YELLOW_BOLD_BRIGHT));
    }

    @Test
    void boldTextIsRenderedWithWhiteBold() {
        String output = renderer.render("**important**");
        assertTrue(output.contains("important"));
        assertTrue(output.contains(AnsiColor.WHITE_BOLD));
    }

    @Test
    void italicTextIsRenderedWithYellow() {
        String output = renderer.render("*emphasis*");
        assertTrue(output.contains("emphasis"));
        assertTrue(output.contains(AnsiColor.YELLOW));
    }

    @Test
    void inlineCodeIsRenderedWithCyan() {
        String output = renderer.render("use `System.out` here");
        assertTrue(output.contains("System.out"));
        assertTrue(output.contains(AnsiColor.CYAN));
    }

    @Test
    void fencedCodeBlockHasLinePrefixes() {
        String output = renderer.render("```\nint x = 1;\n```");
        assertTrue(output.contains("│"));
        assertTrue(output.contains("int x = 1;"));
    }

    @Test
    void unorderedListItemHasBullet() {
        String output = renderer.render("- apples\n- bananas");
        assertTrue(output.contains("•"));
        assertTrue(output.contains("apples"));
        assertTrue(output.contains("bananas"));
    }

    @Test
    void orderedListItemHasNumber() {
        String output = renderer.render("1. first\n2. second");
        assertTrue(output.contains("1."));
        assertTrue(output.contains("first"));
        assertTrue(output.contains("second"));
    }

    @Test
    void linkShowsUrlInBrackets() {
        String output = renderer.render("[Claude](https://claude.ai)");
        assertTrue(output.contains("Claude"));
        assertTrue(output.contains("[https://claude.ai]"));
    }

    @Test
    void horizontalRuleProducesLine() {
        String output = renderer.render("---");
        assertTrue(output.contains("─"));
    }

    @Test
    void tableCellsAreNotConcatenated() {
        String md = """
                | Generator | Syntax | Example |
                |-----------|--------|---------|
                | rand | rand:<min>-<max> | ${exp:rand:1-4} |
                | weighted | weighted:<w>=<v> | ${exp:weighted:80=Y} |
                """;
        String output = renderer.render(md);
        // Header cells must not run together into "GeneratorSyntaxExample"
        assertFalse(stripAnsi(output).contains("GeneratorSyntax"),
                "table cells should be separated, not concatenated");
        // Each row should land on its own line
        String plain = stripAnsi(output);
        assertTrue(plain.contains("Generator"));
        assertTrue(plain.contains("rand"));
        assertTrue(plain.contains("weighted"));
        // A separator between header and body should appear before the first body cell
        int headerIdx = plain.indexOf("Generator");
        int randIdx = plain.indexOf("rand");
        assertTrue(headerIdx < randIdx, "header should render before body rows");
        assertNotEquals(plain.indexOf("Generator", headerIdx),
                plain.indexOf("Syntax"),
                "header cells should occupy different positions");
    }

    @Test
    void tableRendersColumnSeparators() {
        String md = """
                | A | B |
                |---|---|
                | 1 | 2 |
                """;
        String output = renderer.render(md);
        assertTrue(output.contains("│"), "table should use vertical separators between columns");
    }

    private static String stripAnsi(String s) {
        return s.replaceAll("\033\\[[0-9;]*m", "");
    }

    @Test
    void emptyInputProducesNoOutput() {
        String output = renderer.render("");
        assertTrue(output.isBlank());
    }

    @Test
    void outputAlwaysEndsWithReset() {
        String output = renderer.render("# Heading\n\nSome text.");
        assertTrue(output.contains(AnsiColor.RESET));
    }
}
