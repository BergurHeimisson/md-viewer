package com.mdviewer;

import org.commonmark.ext.autolink.AutolinkExtension;
import org.commonmark.ext.gfm.strikethrough.Strikethrough;
import org.commonmark.ext.gfm.strikethrough.StrikethroughExtension;
import org.commonmark.ext.gfm.tables.TableBlock;
import org.commonmark.ext.gfm.tables.TableBody;
import org.commonmark.ext.gfm.tables.TableCell;
import org.commonmark.ext.gfm.tables.TableHead;
import org.commonmark.ext.gfm.tables.TableRow;
import org.commonmark.ext.gfm.tables.TablesExtension;
import org.commonmark.ext.heading.anchor.HeadingAnchorExtension;
import org.commonmark.ext.task.list.items.TaskListItemsExtension;
import org.commonmark.node.*;
import org.commonmark.parser.Parser;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class TerminalRenderer {

    private static final List<org.commonmark.Extension> EXTENSIONS = Arrays.asList(
            TablesExtension.create(),
            StrikethroughExtension.create(),
            TaskListItemsExtension.create(),
            AutolinkExtension.create(),
            HeadingAnchorExtension.create()
    );

    private static final Parser PARSER = Parser.builder()
            .extensions(EXTENSIONS)
            .build();

    public String render(String markdown) {
        Node document = PARSER.parse(markdown);
        AnsiVisitor visitor = new AnsiVisitor();
        document.accept(visitor);
        return visitor.getResult();
    }

    private static class AnsiVisitor extends AbstractVisitor {

        private final StringBuilder sb = new StringBuilder();
        private int orderedListCounter = 0;

        @Override
        public void visit(Heading heading) {
            switch (heading.getLevel()) {
                case 1 -> {
                    sb.append(AnsiColor.CYAN_BOLD).append("═══ ");
                    visitChildren(heading);
                    sb.append(AnsiColor.RESET).append("\n\n");
                }
                case 2 -> {
                    sb.append(AnsiColor.YELLOW_BOLD_BRIGHT).append("─── ");
                    visitChildren(heading);
                    sb.append(AnsiColor.RESET).append("\n\n");
                }
                default -> {
                    sb.append(AnsiColor.PURPLE_BOLD).append("▸ ");
                    visitChildren(heading);
                    sb.append(AnsiColor.RESET).append("\n\n");
                }
            }
        }

        @Override
        public void visit(StrongEmphasis strongEmphasis) {
            sb.append(AnsiColor.WHITE_BOLD);
            visitChildren(strongEmphasis);
            sb.append(AnsiColor.RESET);
        }

        @Override
        public void visit(Emphasis emphasis) {
            sb.append(AnsiColor.YELLOW);
            visitChildren(emphasis);
            sb.append(AnsiColor.RESET);
        }

        @Override
        public void visit(Code code) {
            sb.append(AnsiColor.CYAN).append(code.getLiteral()).append(AnsiColor.RESET);
        }

        @Override
        public void visit(FencedCodeBlock fencedCodeBlock) {
            sb.append("\n");
            for (String line : fencedCodeBlock.getLiteral().split("\n", -1)) {
                sb.append(AnsiColor.CYAN).append("  │ ").append(line).append(AnsiColor.RESET).append("\n");
            }
            sb.append("\n");
        }

        @Override
        public void visit(IndentedCodeBlock indentedCodeBlock) {
            sb.append("\n");
            for (String line : indentedCodeBlock.getLiteral().split("\n", -1)) {
                sb.append(AnsiColor.CYAN).append("  │ ").append(line).append(AnsiColor.RESET).append("\n");
            }
            sb.append("\n");
        }

        @Override
        public void visit(Link link) {
            visitChildren(link);
            sb.append(AnsiColor.BLUE).append(" [").append(link.getDestination()).append("]").append(AnsiColor.RESET);
        }

        @Override
        public void visit(BulletList bulletList) {
            visitChildren(bulletList);
            sb.append("\n");
        }

        @Override
        public void visit(OrderedList orderedList) {
            int saved = orderedListCounter;
            orderedListCounter = orderedList.getMarkerStartNumber();
            visitChildren(orderedList);
            orderedListCounter = saved;
            sb.append("\n");
        }

        @Override
        public void visit(ListItem listItem) {
            Node parent = listItem.getParent();
            if (parent instanceof OrderedList) {
                sb.append("  ").append(orderedListCounter++).append(". ");
            } else {
                sb.append("  ").append(AnsiColor.CYAN).append("•").append(AnsiColor.RESET).append(" ");
            }
            visitChildren(listItem);
        }

        @Override
        public void visit(BlockQuote blockQuote) {
            sb.append(AnsiColor.GREEN_BOLD).append("▌ ");
            visitChildren(blockQuote);
            sb.append(AnsiColor.RESET).append("\n");
        }

        @Override
        public void visit(ThematicBreak thematicBreak) {
            sb.append(AnsiColor.BLACK_BRIGHT).append("─".repeat(60)).append(AnsiColor.RESET).append("\n\n");
        }

        @Override
        public void visit(Paragraph paragraph) {
            visitChildren(paragraph);
            sb.append("\n\n");
        }

        @Override
        public void visit(SoftLineBreak softLineBreak) {
            sb.append(" ");
        }

        @Override
        public void visit(HardLineBreak hardLineBreak) {
            sb.append("\n");
        }

        @Override
        public void visit(CustomNode customNode) {
            if (customNode instanceof Strikethrough) {
                sb.append(AnsiColor.BLACK_BRIGHT);
                visitChildren(customNode);
                sb.append(AnsiColor.RESET);
            } else {
                visitChildren(customNode);
            }
        }

        @Override
        public void visit(CustomBlock customBlock) {
            if (customBlock instanceof TableBlock) {
                renderTable((TableBlock) customBlock);
            } else {
                visitChildren(customBlock);
            }
        }

        /** Renders a GFM table as a box-drawn grid with aligned columns. */
        private void renderTable(TableBlock table) {
            List<List<String>> rows = new ArrayList<>();
            int headerRowCount = 0;

            for (Node section = table.getFirstChild(); section != null; section = section.getNext()) {
                boolean isHead = section instanceof TableHead;
                for (Node rowNode = section.getFirstChild(); rowNode != null; rowNode = rowNode.getNext()) {
                    if (!(rowNode instanceof TableRow)) {
                        continue;
                    }
                    List<String> cells = new ArrayList<>();
                    for (Node cellNode = rowNode.getFirstChild(); cellNode != null; cellNode = cellNode.getNext()) {
                        if (cellNode instanceof TableCell) {
                            cells.add(renderInline(cellNode));
                        }
                    }
                    rows.add(cells);
                    if (isHead) {
                        headerRowCount++;
                    }
                }
            }

            if (rows.isEmpty()) {
                return;
            }

            int columns = rows.stream().mapToInt(List::size).max().orElse(0);
            int[] widths = new int[columns];
            for (List<String> row : rows) {
                for (int c = 0; c < row.size(); c++) {
                    widths[c] = Math.max(widths[c], visibleLength(row.get(c)));
                }
            }

            sb.append("\n");
            appendBorder(widths, "┌", "┬", "┐");
            for (int r = 0; r < rows.size(); r++) {
                appendRow(rows.get(r), widths, r < headerRowCount);
                if (r == headerRowCount - 1) {
                    appendBorder(widths, "├", "┼", "┤");
                }
            }
            appendBorder(widths, "└", "┴", "┘");
            sb.append("\n");
        }

        private void appendBorder(int[] widths, String left, String mid, String right) {
            sb.append(AnsiColor.BLACK_BRIGHT).append(left);
            for (int c = 0; c < widths.length; c++) {
                sb.append("─".repeat(widths[c] + 2));
                sb.append(c == widths.length - 1 ? right : mid);
            }
            sb.append(AnsiColor.RESET).append("\n");
        }

        private void appendRow(List<String> cells, int[] widths, boolean header) {
            for (int c = 0; c < widths.length; c++) {
                sb.append(AnsiColor.BLACK_BRIGHT).append("│").append(AnsiColor.RESET).append(" ");
                String content = c < cells.size() ? cells.get(c) : "";
                if (header) {
                    sb.append(AnsiColor.WHITE_BOLD).append(content).append(AnsiColor.RESET);
                } else {
                    sb.append(content);
                }
                sb.append(" ".repeat(widths[c] - visibleLength(content))).append(" ");
            }
            sb.append(AnsiColor.BLACK_BRIGHT).append("│").append(AnsiColor.RESET).append("\n");
        }

        /** Renders a node's inline children to ANSI, trimmed to a single line. */
        private String renderInline(Node node) {
            AnsiVisitor sub = new AnsiVisitor();
            for (Node child = node.getFirstChild(); child != null; child = child.getNext()) {
                child.accept(sub);
            }
            return sub.getResult().strip();
        }

        /** Visible character count, ignoring ANSI escape sequences. */
        private static int visibleLength(String s) {
            return s.replaceAll("\033\\[[0-9;]*m", "").length();
        }

        @Override
        public void visit(Text text) {
            sb.append(text.getLiteral());
        }

        @Override
        public void visit(HtmlInline htmlInline) {
            // Preserve raw markup literally (e.g. <min>, <max>) instead of dropping it.
            sb.append(htmlInline.getLiteral());
        }

        @Override
        public void visit(HtmlBlock htmlBlock) {
            sb.append(htmlBlock.getLiteral());
            sb.append("\n");
        }

        public String getResult() {
            return sb.toString();
        }
    }
}
