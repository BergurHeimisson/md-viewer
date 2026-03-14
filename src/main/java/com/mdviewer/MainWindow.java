package com.mdviewer;

import com.formdev.flatlaf.util.SystemFileChooser;
import javax.swing.*;
import javax.swing.event.HyperlinkEvent;
import javax.swing.text.html.HTMLEditorKit;
import javax.swing.text.html.StyleSheet;
import java.awt.*;
import java.awt.datatransfer.DataFlavor;
import java.awt.dnd.*;
import java.awt.event.*;
import javax.swing.AbstractAction;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.prefs.Preferences;

/**
 * The single application window.
 *
 * Layout (BorderLayout):
 *   CENTER — JScrollPane wrapping a JEditorPane
 *
 * No toolbar, no menu bar — just the scrollable content area.
 * Drag-and-drop is registered on the editor pane so the entire
 * window surface acts as a drop target.
 */
public class MainWindow extends JFrame {

    // Default window dimensions when no saved preferences exist
    private static final int DEFAULT_WIDTH  = 900;
    private static final int DEFAULT_HEIGHT = 680;

    private static final String APP_TITLE = "mdViewer";

    private final Preferences prefs;
    private final MarkdownRenderer mdRenderer;
    private final JEditorPane editorPane;
    private final JScrollPane scrollPane;

    /** Currently loaded file (may be null). */
    private File currentFile;

    /** Sorted list of .md/.markdown siblings in the same directory. */
    private List<File> siblings = new ArrayList<>();

    public MainWindow(Preferences prefs) {
        super(APP_TITLE);
        this.prefs      = prefs;
        this.mdRenderer = new MarkdownRenderer();

        // ------------------------------------------------------------------ //
        // Editor pane — displays rendered HTML
        // ------------------------------------------------------------------ //
        editorPane = new JEditorPane();
        editorPane.setEditable(false);
        editorPane.setContentType("text/html");

        // Override the default StyleSheet so JEditorPane does not inject its
        // own opinionated defaults (blue links, serif fonts, etc.)
        HTMLEditorKit kit = new HTMLEditorKit();
        StyleSheet ss = new StyleSheet();          // blank sheet — CSS comes from the HTML
        kit.setStyleSheet(ss);
        editorPane.setEditorKit(kit);

        // Match editor background to Dracula so there is no flash of white
        editorPane.setBackground(App.colors.awtBg());
        editorPane.setForeground(App.colors.awtFg());

        // Open hyperlinks in the system browser
        editorPane.addHyperlinkListener(e -> {
            if (e.getEventType() == HyperlinkEvent.EventType.ACTIVATED && e.getURL() != null) {
                try {
                    Desktop.getDesktop().browse(e.getURL().toURI());
                } catch (Exception ex) {
                    // Silently ignore — non-critical
                }
            }
        });

        // ------------------------------------------------------------------ //
        // Scroll pane
        // ------------------------------------------------------------------ //
        scrollPane = new JScrollPane(editorPane);
        scrollPane.setBorder(BorderFactory.createEmptyBorder());
        scrollPane.setHorizontalScrollBarPolicy(ScrollPaneConstants.HORIZONTAL_SCROLLBAR_AS_NEEDED);
        scrollPane.setVerticalScrollBarPolicy(ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS);
        scrollPane.getVerticalScrollBar().setUnitIncrement(16);

        // ------------------------------------------------------------------ //
        // Root panel — gives us a coloured background behind the scroll pane
        // ------------------------------------------------------------------ //
        JPanel root = new JPanel(new BorderLayout());
        root.setBackground(App.colors.awtBg());
        root.add(scrollPane, BorderLayout.CENTER);
        setContentPane(root);

        // ------------------------------------------------------------------ //
        // Window setup
        // ------------------------------------------------------------------ //
        setDefaultCloseOperation(JFrame.DO_NOTHING_ON_CLOSE);
        addWindowListener(new WindowAdapter() {
            @Override
            public void windowClosing(WindowEvent e) {
                savePreferences();
                dispose();
                System.exit(0);
            }
        });

        restorePreferences();
        showPlaceholder();
        registerDropTarget();
        registerKeyBindings();
    }

    // ======================================================================= //
    // Public API
    // ======================================================================= //

    /**
     * Reads and renders the given Markdown file.
     * Safe to call from any thread — marshals to the EDT internally.
     */
    public void openFile(File file) {
        final File absFile = file.getAbsoluteFile();
        if (!absFile.isFile()) return;

        SwingUtilities.invokeLater(() -> {
            try {
                String markdown = Files.readString(absFile.toPath(), StandardCharsets.UTF_8);
                String html     = mdRenderer.render(markdown);

                currentFile = absFile;
                rebuildSiblings(absFile);
                editorPane.setText(html);

                // Scroll back to the top after loading new content
                SwingUtilities.invokeLater(() ->
                        editorPane.setCaretPosition(0));

                setTitle(absFile.getName() + " — " + APP_TITLE);

            } catch (IOException ex) {
                JOptionPane.showMessageDialog(
                        this,
                        "Could not read file:\n" + ex.getMessage(),
                        "Error",
                        JOptionPane.ERROR_MESSAGE
                );
            }
        });
    }

    // ======================================================================= //
    // Sibling navigation
    // ======================================================================= //

    private void rebuildSiblings(File file) {
        File dir = file.getAbsoluteFile().getParentFile();
        File[] all = dir.listFiles(f -> {
            String n = f.getName().toLowerCase();
            return f.isFile() && (n.endsWith(".md") || n.endsWith(".markdown"));
        });
        siblings = all != null ? new ArrayList<>(Arrays.asList(all)) : new ArrayList<>();
        siblings.sort((a, b) -> a.getName().compareToIgnoreCase(b.getName()));
    }

    private void navigateSibling(int delta) {
        if (currentFile == null || siblings.isEmpty()) return;
        int idx = siblings.indexOf(currentFile);
        if (idx < 0) return;
        int next = (idx + delta + siblings.size()) % siblings.size();
        openFile(siblings.get(next));
    }

    // ======================================================================= //
    // Placeholder
    // ======================================================================= //

    private void showPlaceholder() {
        // Build a centred placeholder directly as HTML matching the Dracula theme
        String html = "<!DOCTYPE html><html><head><style>"
                + "body { background:" + App.colors.bg + "; margin:0; padding:0; height:100%; }"
                + "div.wrap { display:table; width:100%; height:100%; }"
                + "div.cell { display:table-cell; text-align:center; vertical-align:middle; }"
                + "p.icon { font-size:64px; margin:0 0 16px; }"
                + "p.main { color:" + App.colors.fg + "; font-family:sans-serif; "
                +          "font-size:22px; margin:0 0 10px; font-weight:600; }"
                + "p.sub  { color:" + App.colors.muted + "; font-family:sans-serif; "
                +          "font-size:14px; margin:0; }"
                + "</style></head><body>"
                + "<div class='wrap'><div class='cell'>"
                + "<p class='icon'>&#128196;</p>"
                + "<p class='main'>Drop a Markdown file here</p>"
                + "<p class='sub'>or launch with: mdviewer path/to/file.md</p>"
                + "</div></div></body></html>";

        editorPane.setText(html);
        setTitle(APP_TITLE);
        currentFile = null;
    }

    // ======================================================================= //
    // Drag and drop
    // ======================================================================= //

    private void registerDropTarget() {
        // Make the whole editor pane (= visible window surface) a drop target
        new DropTarget(editorPane, DnDConstants.ACTION_COPY_OR_MOVE,
                new DropTargetAdapter() {
                    @Override
                    public void dragEnter(DropTargetDragEvent dtde) {
                        if (isMarkdownDrop(dtde.getCurrentDataFlavors())) {
                            dtde.acceptDrag(DnDConstants.ACTION_COPY);
                            // Subtle visual feedback — slightly highlight the border
                            scrollPane.setBorder(BorderFactory.createLineBorder(
                                    App.colors.awtHeading(), 2));
                        } else {
                            dtde.rejectDrag();
                        }
                    }

                    @Override
                    public void dragExit(DropTargetEvent dte) {
                        scrollPane.setBorder(BorderFactory.createEmptyBorder());
                    }

                    @Override
                    public void drop(DropTargetDropEvent dtde) {
                        scrollPane.setBorder(BorderFactory.createEmptyBorder());
                        try {
                            dtde.acceptDrop(DnDConstants.ACTION_COPY);
                            @SuppressWarnings("unchecked")
                            List<File> files = (List<File>)
                                    dtde.getTransferable()
                                        .getTransferData(DataFlavor.javaFileListFlavor);
                            if (!files.isEmpty()) {
                                File dropped = files.get(0);
                                if (dropped.getName().toLowerCase().endsWith(".md")
                                        || dropped.getName().toLowerCase().endsWith(".markdown")) {
                                    openFile(dropped);
                                } else {
                                    JOptionPane.showMessageDialog(
                                            MainWindow.this,
                                            "Only .md / .markdown files are supported.",
                                            "Unsupported file type",
                                            JOptionPane.WARNING_MESSAGE
                                    );
                                }
                            }
                            dtde.dropComplete(true);
                        } catch (Exception ex) {
                            dtde.dropComplete(false);
                        }
                    }
                }, true);
    }

    private boolean isMarkdownDrop(DataFlavor[] flavors) {
        for (DataFlavor f : flavors) {
            if (DataFlavor.javaFileListFlavor.equals(f)) return true;
        }
        return false;
    }

    // ======================================================================= //
    // Key bindings
    // ======================================================================= //

    private void registerKeyBindings() {
        // Cmd/Ctrl + W or Escape to quit
        KeyStroke cmdW = KeyStroke.getKeyStroke(KeyEvent.VK_W,
                Toolkit.getDefaultToolkit().getMenuShortcutKeyMaskEx());
        KeyStroke escape = KeyStroke.getKeyStroke(KeyEvent.VK_ESCAPE, 0);

        ActionListener quit = e -> {
            savePreferences();
            dispose();
            System.exit(0);
        };

        getRootPane().registerKeyboardAction(quit, cmdW,
                JComponent.WHEN_IN_FOCUSED_WINDOW);
        getRootPane().registerKeyboardAction(quit, escape,
                JComponent.WHEN_IN_FOCUSED_WINDOW);

        // Cmd/Ctrl + O — open file via dialog
        KeyStroke cmdO = KeyStroke.getKeyStroke(KeyEvent.VK_O,
                Toolkit.getDefaultToolkit().getMenuShortcutKeyMaskEx());
        getRootPane().registerKeyboardAction(e -> openFileDialog(), cmdO,
                JComponent.WHEN_IN_FOCUSED_WINDOW);

        // Cmd/Ctrl + R — reload current file
        KeyStroke cmdR = KeyStroke.getKeyStroke(KeyEvent.VK_R,
                Toolkit.getDefaultToolkit().getMenuShortcutKeyMaskEx());
        getRootPane().registerKeyboardAction(e -> {
            if (currentFile != null) openFile(currentFile);
        }, cmdR, JComponent.WHEN_IN_FOCUSED_WINDOW);

        // Left / Right arrow — cycle through sibling .md files.
        // Must be bound on the editorPane itself (WHEN_FOCUSED) because JEditorPane
        // consumes arrow keys for cursor movement before they reach the root pane.
        InputMap im = editorPane.getInputMap(JComponent.WHEN_FOCUSED);
        ActionMap am = editorPane.getActionMap();
        im.put(KeyStroke.getKeyStroke(KeyEvent.VK_LEFT, 0), "nav-prev");
        im.put(KeyStroke.getKeyStroke(KeyEvent.VK_RIGHT, 0), "nav-next");
        am.put("nav-prev", new AbstractAction() {
            public void actionPerformed(ActionEvent e) { navigateSibling(-1); }
        });
        am.put("nav-next", new AbstractAction() {
            public void actionPerformed(ActionEvent e) { navigateSibling(+1); }
        });

        // Up / Down arrow — scroll the viewport
        im.put(KeyStroke.getKeyStroke(KeyEvent.VK_UP, 0), "scroll-up");
        im.put(KeyStroke.getKeyStroke(KeyEvent.VK_DOWN, 0), "scroll-down");
        am.put("scroll-up", new AbstractAction() {
            public void actionPerformed(ActionEvent e) {
                JScrollBar bar = scrollPane.getVerticalScrollBar();
                bar.setValue(bar.getValue() - bar.getUnitIncrement());
            }
        });
        am.put("scroll-down", new AbstractAction() {
            public void actionPerformed(ActionEvent e) {
                JScrollBar bar = scrollPane.getVerticalScrollBar();
                bar.setValue(bar.getValue() + bar.getUnitIncrement());
            }
        });
    }

    private void openFileDialog() {
        SystemFileChooser chooser = new SystemFileChooser();
        chooser.setDialogTitle("Open Markdown File");
        chooser.setFileFilter(new SystemFileChooser.FileNameExtensionFilter(
                "Markdown files (*.md, *.markdown)", "md", "markdown"));
        if (currentFile != null) {
            chooser.setCurrentDirectory(currentFile.getParentFile());
        }
        if (chooser.showOpenDialog(this) == SystemFileChooser.APPROVE_OPTION) {
            openFile(chooser.getSelectedFile());
        }
    }

    // ======================================================================= //
    // Preferences — window size/position
    // ======================================================================= //

    private void restorePreferences() {
        int x = prefs.getInt(App.PREF_X, -1);
        int y = prefs.getInt(App.PREF_Y, -1);
        int w = prefs.getInt(App.PREF_WIDTH,  DEFAULT_WIDTH);
        int h = prefs.getInt(App.PREF_HEIGHT, DEFAULT_HEIGHT);

        setSize(w, h);

        if (x < 0 || y < 0) {
            // Centre on screen when no saved position exists
            setLocationRelativeTo(null);
        } else {
            setLocation(x, y);
            // Guard against saved position being off-screen (e.g., after monitor change)
            ensureOnScreen();
        }
    }

    private void savePreferences() {
        prefs.putInt(App.PREF_X,      getX());
        prefs.putInt(App.PREF_Y,      getY());
        prefs.putInt(App.PREF_WIDTH,  getWidth());
        prefs.putInt(App.PREF_HEIGHT, getHeight());
        try {
            prefs.flush();
        } catch (Exception ignored) {}
    }

    /**
     * Moves the window back onto a visible screen area if it has drifted off.
     */
    private void ensureOnScreen() {
        GraphicsEnvironment ge = GraphicsEnvironment.getLocalGraphicsEnvironment();
        Rectangle screen = ge.getMaximumWindowBounds();

        int x = Math.max(screen.x, Math.min(getX(), screen.x + screen.width  - getWidth()));
        int y = Math.max(screen.y, Math.min(getY(), screen.y + screen.height - getHeight()));
        setLocation(x, y);
    }
}
