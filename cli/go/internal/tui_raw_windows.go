//go:build windows

package core

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func (t *TUI) enterRawMode() error {
	if t.consoleModeSet {
		return nil
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle := syscall.Handle(os.Stdin.Fd())

	ret, _, err := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&t.origConsoleMode)))
	if ret == 0 {
		return fmt.Errorf("GetConsoleMode: %v", err)
	}

	// Disable ENABLE_LINE_INPUT, ENABLE_ECHO_INPUT, ENABLE_PROCESSED_INPUT
	// Enable ENABLE_VIRTUAL_TERMINAL_INPUT
	newMode := t.origConsoleMode & ^uint32(0x0001|0x0002|0x0004)
	newMode |= 0x0200

	ret, _, err = setConsoleMode.Call(uintptr(handle), uintptr(newMode))
	if ret == 0 {
		return fmt.Errorf("SetConsoleMode: %v", err)
	}

	t.consoleModeSet = true
	return nil
}

func (t *TUI) exitRawMode() {
	if !t.consoleModeSet {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	handle := syscall.Handle(os.Stdin.Fd())
	setConsoleMode.Call(uintptr(handle), uintptr(t.origConsoleMode))
	t.consoleModeSet = false
}
