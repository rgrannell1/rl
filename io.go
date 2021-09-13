package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/smallnest/ringbuffer"
)

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

// Read input from stdin into a circular-buffer
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

// Substitute user-input into a command in place of the environment name;
// useful to visualise what was run by the user
func SubstitueCommand(execute *string, input *string) string {
	return strings.ReplaceAll(*execute, "$"+ENVAR_NAME_RL_INPUT, *input)
}

// Get line-count in bytes.Buffer
func LineCounter(content *bytes.Buffer) int {
	count := 0

	for {
		_, err := content.ReadBytes('\n')

		if err == nil {
			count += 1
		} else {
			return count
		}
	}
}
