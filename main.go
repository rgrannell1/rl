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

Options:
	--clear    clear the terminal after each update.
	--empty    allow empty.
`
	opts, _ := docopt.ParseDoc(usage)
	clear, err := opts.Bool("--clear")

	if err != nil {
		fmt.Printf("RL: failed to read clear option. %v\n", err)
		os.Exit(1)
	}

	rl(clear)
}
