package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/docopt/docopt-go"
	"github.com/eiannone/keyboard"
)

// Clear the terminal by writing  an ansi control sequence to /dev/tty
func clearTTY() error {
	fd, err := syscall.Open("/dev/tty", syscall.O_WRONLY, 00200)

	if err != nil {
		return err
	}

	file := os.NewFile(uintptr(fd), "pipe")
	defer func() {
		file.Close()
	}()

	if _, writeErr := file.Write([]byte("\033[H\033[2J")); writeErr != nil {
		return writeErr
	}

	return nil
}

// Start the interactive line-editor
func rl(clear bool) {
	if err := keyboard.Open(); err != nil {
		fmt.Printf("RL: failed to read from keyboard. %v\n", err)
		os.Exit(1)
	}
	defer func() {
		keyboard.Close()
	}()

	linebuffer := []rune{}

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
			break
		} else if key == keyboard.KeySpace {
			// handle spaces
			linebuffer = append(linebuffer, ' ')
		} else {
			// -- append character to the buffer
			linebuffer = append(linebuffer, char)
		}

		// print non-empty buffers
		if len(linebuffer) > 0 {
			if clear {
				// we don't really care enough about clear errors to spam the console; ignore the errors.
				clearTTY()
			}

			fmt.Println(string(linebuffer))
		}
	}
}

func main() {
	usage := `rl
Usage:
	rl [--clear] [--empty]
Description:
  rl (readline) is a minimal line editor. It allows users to run grep, ls, and other commands in a pseudo-interactive mode,
	where a user enters a filter and commands are re-executed live.

	It captures keypresses and backspaces (to remove characters), and terminates when
	escape, enter, or ctrl + c is pressed. By default, it echos its line buffer for each keypress, as shown below:

	❯ rl
	h
	he
	hel
	hell
	hello

	setting --clear will prompt rl to clear the terminal (/dev/tty) after each keypress.

Options:
	--clear    clear the terminal after each update.

License:
	The MIT License

	Copyright (c) 2021 Róisín Grannell

	Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`
	opts, _ := docopt.ParseDoc(usage)
	clear, err := opts.Bool("--clear")

	if err != nil {
		fmt.Printf("RL: failed to read clear option. %v\n", err)
		os.Exit(1)
	}

	rl(clear)
}
