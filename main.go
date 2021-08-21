package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/docopt/docopt-go"
	"github.com/eiannone/keyboard"
)

// Open /dev/tty
func openTTY() (*os.File, error) {
	fd, err := syscall.Open("/dev/tty", syscall.O_WRONLY, 00200)

	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "pipe"), nil
}

// When a line-editor is updated, either emit the line-buffer content or execute a command.
func oneLineChange(state LineChangeState) error {
	// print non-empty buffers
	if len(*state.linebuffer) > 0 {
		if state.clear {
			// we don't really care enough about clear errors to spam the console; ignore the errors,
			// and attempt to clear the TTY. This program should always be run in interactive mode.

			state.tty.Write([]byte("\033[H\033[2J"))
		}

		line := string(*state.linebuffer)
		if len(*state.execute) > 0 {
			// get shell; hopefully the export syntax is right!

			// run the provided command in shell
			toExec := fmt.Sprintf("export RL_INPUT=\"%s\"; %s", line, *state.execute)
			shellCmd := exec.Command(state.shell, "-c", toExec)

			// might want to output to /dev/tty, and only emit the final selection to stdout

			if state.done {
				shellCmd.Stdout = os.Stdout
			} else {
				shellCmd.Stdout = state.tty
			}

			shellCmd.Stderr = os.Stderr

			// handle errors
			shellCmd.Run()

		} else {
			// print out the line buffer
			fmt.Println(line)
		}
	}

	return nil
}

type LineChangeState struct {
	shell      string
	tty        *os.File
	linebuffer *[]rune
	clear      bool
	execute    *string
	done       bool
}

// Start the interactive line-editor
func rl(clear bool, execute *string) {
	if err := keyboard.Open(); err != nil {
		fmt.Printf("RL: failed to read from keyboard. %v\n", err)
		os.Exit(1)
	}
	defer func() {
		keyboard.Close()
	}()

	linebuffer := []rune{}

	tty, ttyErr := openTTY()

	if ttyErr != nil {
		fmt.Printf("RL: failed to open /dev/tty. %v\n", ttyErr)
		os.Exit(1)
	}
	defer func() {
		tty.Close()
	}()

	state := LineChangeState{
		os.Getenv("SHELL"),
		tty,
		&linebuffer,
		clear,
		execute,
		false,
	}

	for {
		// repeatedly get keys, until a terminating character is reached
		char, key, err := keyboard.GetKey()

		if err != nil {
			fmt.Printf("RL: Keyboard read failed. %v\n", err)
			os.Exit(1)
		}

		if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			if len(linebuffer) > 0 {
				// backspace should remove the last element in a buffer
				linebuffer = linebuffer[:len(linebuffer)-1]
			}
		} else if key == keyboard.KeyCtrlC || key == keyboard.KeyEsc || key == keyboard.KeyEnter {
			// exit interactive editor

			state.done = true
		} else if key == keyboard.KeySpace {
			// handle spaces
			linebuffer = append(linebuffer, ' ')
		} else {
			// -- append character to the buffer
			linebuffer = append(linebuffer, char)
		}

		// -- it might be nice to have this as a goroutine with a mutex lock
		oneLineChange(state)

		if state.done {
			break
		}
	}
}

func main() {
	usage := `rl
Usage:
	rl [-c|--clear] [-x <cmd>|--execute <cmd>]
Description:
  rl (readline) is a minimal line editor. It allows users to run grep, ls, and other commands in a pseudo-interactive mode,z
	where a user enters a filter and commands are re-executed live.

	It captures keypresses and backspaces (to remove characters), and terminates when
	escape, enter, or ctrl + c is pressed. By default, it echos its line buffer for each keypress, as shown below:

	> rl

	h
	he
	hel
	hell
	hello

	setting --clear will prompt rl to clear the terminal (/dev/tty) after each keypress.

Options:
	-c, --clear                          clear the terminal after each update.
	-x <command>, --execute <command>    execute a command on readline change; the current line will be available as the line $RL_INPUT

License:
	The MIT License

	Copyright (c) 2021 Róisín Grannell

	Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`
	opts, _ := docopt.ParseDoc(usage)
	clear, clearErr := opts.Bool("--clear")

	if clearErr != nil {
		fmt.Printf("RL: failed to read clear option. %v\n", clearErr)
		os.Exit(1)
	}

	execute, execErr := opts.String("--execute")

	if execErr != nil {
		execute = ""
	}

	rl(clear, &execute)
}
