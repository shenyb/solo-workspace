//go:build !windows

package core

import (
	"os"

	"golang.org/x/term"
)

var savedTermState *term.State

func enterRawMode() error {
	if savedTermState != nil {
		return nil
	}
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	savedTermState = state
	return nil
}

func exitRawMode() {
	if savedTermState == nil {
		return
	}
	fd := int(os.Stdin.Fd())
	term.Restore(fd, savedTermState)
	savedTermState = nil
}
