package core

import (
	"fmt"
	"os"
	"os/exec"
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

	// Windows raw mode state (used by tui_raw_windows.go)
	origConsoleMode uint32
	consoleModeSet  bool
}

// NewTUI creates a new TUI instance
func NewTUI(title string, items []MenuItem) *TUI {
	return &TUI{
		title: title,
		items: items,
	}
}

// Run displays the TUI and handles input loop
func (t *TUI) Run() error {
	if err := t.enterRawMode(); err != nil {
		return err
	}
	defer t.exitRawMode()

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

// render clears screen and draws the menu
func (t *TUI) render() {
	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	// Title
	fmt.Printf("\033[1;36m%s\033[0m\n\n", t.title)

	// Items
	for i, item := range t.items {
		if i == t.pos {
			fmt.Printf(" \033[7m > %-40s \033[0m  %s\n", item.Label, Dim(item.Description))
		} else {
			fmt.Printf("   %-40s    %s\n", item.Label, Dim(item.Description))
		}
	}

	// Footer
	fmt.Printf("\n\033[2m↑↓ navigate  Enter select  q quit\033[0m\n")
}

// readKey reads a single keypress in raw mode
func (t *TUI) readKey() (string, error) {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return "", err
	}

	// Arrow keys are 3 bytes: \033 [ A/B/C/D
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

	// Single character
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

	// Handle back item
	if item.IsBack {
		t.done = true
		return
	}

	// Handle submenu
	if len(item.Children) > 0 {
		sub := NewTUI(t.title+" > "+item.Label, item.Children)
		if err := sub.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		if sub.CommandRan {
			// Submenu ran a command — exit the parent TUI too
			t.done = true
		}
		if sub.WasExit() {
			t.done = true
			t.exit = true
		}
		return
	}

	// Handle command
	if len(item.Command) > 0 {
		t.done = true
		t.executeCommand(item.Command)
	}
}

// executeCommand runs the given command, inheriting stdin/stdout/stderr
func (t *TUI) executeCommand(args []string) {
	t.CommandRan = true

	t.exitRawMode()

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
	}

	t.enterRawMode()
}
