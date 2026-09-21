package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var originalState syscall.Termios

func initTerminal() {
	fd := int(os.Stdin.Fd())
	// Get current terminal state via raw syscall
	var t syscall.Termios
	syscall.Syscall6(syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&t)),
		0, 0, 0)
	originalState = t

	// Set raw mode
	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	t.Iflag &^= syscall.IXON | syscall.ICRNL
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 0

	syscall.Syscall6(syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(0x5402), // TCSETS
		uintptr(unsafe.Pointer(&t)),
		0, 0, 0)
}

func restoreTerminal() {
	fd := int(os.Stdin.Fd())
	syscall.Syscall6(syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(0x5402), // TCSETS
		uintptr(unsafe.Pointer(&originalState)),
		0, 0, 0)
	fmt.Print("\033[?25h\033[0m\033[2J\033[H")
}

func readKey() (byte, bool) {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return 0, false
	}

	if n == 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65: return 'w', true
		case 66: return 's', true
		case 67: return 'l', true
		case 68: return 'j', true
		}
	}

	return buf[0], true
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func beep() {
	fmt.Print("\a")
}
