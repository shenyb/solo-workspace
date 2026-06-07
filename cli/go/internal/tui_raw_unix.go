//go:build !windows

package core

import (
	"golang.org/x/term"
	"os"
)

var originalTermState *term.State

func (t *TUI) enterRawMode() error {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	originalTermState = state
	return nil
}

func (t *TUI) exitRawMode() {
	if originalTermState != nil {
		fd := int(os.Stdin.Fd())
		term.Restore(fd, originalTermState)
		originalTermState = nil
	}
}
