//go:build windows

package core

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	savedConsoleMode uint32
	consoleModeSaved bool
)

func enterRawMode() error {
	if consoleModeSaved {
		return nil
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle := syscall.Handle(os.Stdin.Fd())

	ret, _, err := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&savedConsoleMode)))
	if ret == 0 {
		return fmt.Errorf("GetConsoleMode: %v", err)
	}

	newMode := savedConsoleMode & ^uint32(0x0001|0x0002|0x0004)
	newMode |= 0x0200

	ret, _, err = setConsoleMode.Call(uintptr(handle), uintptr(newMode))
	if ret == 0 {
		return fmt.Errorf("SetConsoleMode: %v", err)
	}

	consoleModeSaved = true
	return nil
}

func exitRawMode() {
	if !consoleModeSaved {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	handle := syscall.Handle(os.Stdin.Fd())
	setConsoleMode.Call(uintptr(handle), uintptr(savedConsoleMode))
	consoleModeSaved = false
}
