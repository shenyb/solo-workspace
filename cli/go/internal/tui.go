package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-runewidth"
)

// ── TUI Types ─────────────────────────────────────────────

// MenuItem represents a selectable item in the TUI
type MenuItem struct {
	Label       string
	Description string
	Command     []string // exec args to run when selected
	Children    []MenuItem
	IsBack      bool // ".." back item
}

// TUI is a simple text-based interactive menu
type TUI struct {
	title string
	items []MenuItem
	pos   int // current selection position
	done  bool
	exit  bool

	// True when a command was executed (vs. user quitting)
	CommandRan bool

	// Only the root menu manages terminal raw mode; submenus reuse it.
	managesTerminal bool
}

// NewTUI creates a new root TUI instance
func NewTUI(title string, items []MenuItem) *TUI {
	return &TUI{
		title:           title,
		items:           items,
		managesTerminal: true,
	}
}

func newSubTUI(title string, items []MenuItem) *TUI {
	return &TUI{
		title:           title,
		items:           items,
		managesTerminal: false,
	}
}

// Run displays the TUI and handles input loop
func (t *TUI) Run() error {
	if t.managesTerminal {
		if err := enterRawMode(); err != nil {
			return err
		}
		defer func() {
			restoreTerminal(!t.CommandRan && t.exit)
		}()
	}

	for !t.done {
		t.render()
		key, err := t.readKey()
		if err != nil {
			return err
		}
		t.handleKey(key)
	}
	return nil
}

// WasExit returns true if user chose to exit
func (t *TUI) WasExit() bool {
	return t.exit
}

// Dim returns a dimmed (gray) version of the text
func Dim(s string) string {
	if s == "" {
		return ""
	}
	return "\033[2m" + s + "\033[0m"
}

// restoreTerminal exits raw mode and resets terminal display attributes.
func restoreTerminal(clearMenu bool) {
	exitRawMode()
	if clearMenu {
		tuiWriteRaw("\033[2J\033[H")
	}
	tuiWriteRaw("\033[0m\033[?25h")
}

// tuiWriteRaw writes without adding line endings (for escape sequences).
func tuiWriteRaw(s string) {
	_, _ = os.Stdout.WriteString(s)
}

// tuiLine writes one screen line. Raw mode does not map \n to \r\n, so each
// line must start with \r to avoid column drift.
func tuiLine(s string) {
	_, _ = os.Stdout.WriteString("\r" + s + "\r\n")
}

func padDisplayRight(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w >= width {
		return runewidth.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-w)
}

// render clears screen and draws the menu
func (t *TUI) render() {
	tuiWriteRaw("\r\033[2J\033[H")

	tuiLine(fmt.Sprintf("\033[1;36m%s\033[0m", t.title))
	tuiLine("")

	labelWidth := runewidth.StringWidth("Label")
	for _, item := range t.items {
		w := runewidth.StringWidth(item.Label)
		if w > labelWidth {
			labelWidth = w
		}
	}

	colWidth := labelWidth + 2 // room for "> " marker
	header := padDisplayRight("Label", colWidth) + " │ Description"
	tuiLine("\033[2m" + header + "\033[0m")
	tuiLine(strings.Repeat("─", colWidth) + "─┼" + strings.Repeat("─", 40))

	for i, item := range t.items {
		marker := "  "
		if i == t.pos {
			marker = "> "
		}
		labelCol := padDisplayRight(marker+item.Label, colWidth)
		desc := item.Description
		if i != t.pos {
			desc = Dim(desc)
		}

		line := labelCol + " │ " + desc
		if i == t.pos {
			line = "\033[7m" + line + "\033[0m"
		}
		tuiLine(line)
	}

	tuiLine("")
	tuiLine("\033[2m↑↓ navigate  Enter select  q quit\033[0m")
}

// readKey reads a single keypress in raw mode
func (t *TUI) readKey() (string, error) {
	buf := make([]byte, 6)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return "", err
	}

	if n == 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 'A':
			return "up", nil
		case 'B':
			return "down", nil
		case 'C':
			return "right", nil
		case 'D':
			return "left", nil
		}
	}

	if n == 1 {
		return string(buf[0]), nil
	}

	return "", nil
}

// handleKey processes a keypress
func (t *TUI) handleKey(key string) {
	switch key {
	case "q", "Q":
		t.done = true
		t.exit = true
	case "up":
		if t.pos > 0 {
			t.pos--
		}
	case "down":
		if t.pos < len(t.items)-1 {
			t.pos++
		}
	case "\r", "\n":
		t.activate()
	}
}

// activate selects the current item
func (t *TUI) activate() {
	if t.pos < 0 || t.pos >= len(t.items) {
		return
	}
	item := t.items[t.pos]

	if item.IsBack {
		t.done = true
		return
	}

	if len(item.Children) > 0 {
		sub := newSubTUI(t.title+" > "+item.Label, item.Children)
		if err := sub.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		if sub.CommandRan {
			t.done = true
			t.CommandRan = true
		}
		if sub.WasExit() {
			t.done = true
			t.exit = true
		}
		return
	}

	if len(item.Command) > 0 {
		t.done = true
		t.CommandRan = true
		t.executeCommand(item.Command)
	}
}

// executeCommand runs the given command with a normal (cooked) terminal.
func (t *TUI) executeCommand(args []string) {
	restoreTerminal(false)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
	}
}
