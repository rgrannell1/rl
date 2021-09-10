package main

import (
	"bufio"
	"io"
	"os"
	"syscall"

	"github.com/smallnest/ringbuffer"
)

// Allow user writes, no other permissions
const USER_WRITE_OCTAL = 00200

// Open /dev/tty with user write-only permissions. If it fails to open, return
// an error that will indicate this tool is being run in non-interactive mode
func OpenTTY() (*os.File, error) {
	fd, err := syscall.Open("/dev/tty", syscall.O_WRONLY, USER_WRITE_OCTAL)

	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "pipe"), nil
}

// Detect whether input was piped into stdin; ie. if we
// need to receive that input, buffer it, and pass it to subcommands.
func StdinPiped() (bool, error) {
	fi, err := os.Stdin.Stat()

	if err != nil {
		return false, err
	}

	return fi.Mode()&os.ModeCharDevice == 0, nil
}

// Read input from stdin into a buffer
func StdinReader(input *ringbuffer.RingBuffer) {
	in := bufio.NewReader(os.Stdin)
	for {
		by, err := in.ReadByte()
		if err == io.EOF {
			break
		}

		input.Write([]byte{by})
	}
}
