package com.mdviewer;

import org.commonmark.ext.autolink.AutolinkExtension;
import org.commonmark.ext.gfm.strikethrough.Strikethrough;
import org.commonmark.ext.gfm.strikethrough.StrikethroughExtension;
import org.commonmark.ext.gfm.tables.TablesExtension;
import org.commonmark.ext.heading.anchor.HeadingAnchorExtension;
import org.commonmark.ext.task.list.items.TaskListItemsExtension;
import org.commonmark.node.*;
import org.commonmark.parser.Parser;

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
                    sb.append(AnsiColor.PURPLE_BOLD);
                    sb.append("### ");
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
        public void visit(Text text) {
            sb.append(text.getLiteral());
        }

        public String getResult() {
            return sb.toString();
        }
    }
}
