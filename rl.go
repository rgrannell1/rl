package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

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
		if !state.show {
			// we don't really care enough about clear errors to spam the console; ignore the errors,
			// and attempt to clear the TTY. This program should always be run in interactive mode.

			state.tty.Write([]byte("\033[H\033[2J"))
		}

		line := string(*state.linebuffer)

		if len(*state.execute) > 0 {
			// execute a command

			if state.done && state.inputOnly {
				// we're done, we only want the input line but not the command
				os.Stdout.WriteString(line)
				os.Stdout.WriteString("\n")

				return nil
			}

			os.Setenv("RL_INPUT", *state.execute)

			// run the provided command in the users shell
			//toExec := fmt.Sprintf("export RL_INPUT=\"%s\"; %s", line, *state.execute)
			toExec := fmt.Sprintf(*state.execute)
			cmd := exec.Command(state.shell, "-c", toExec)

			// by default, go will use the current  process's environment. Add RL_INPUT to the list.
			cmd.Env = append(state.environment, "RL_INPUT="+line)

			// only output the last command to standard-output by default
			if state.done {
				cmd.Stdout = os.Stdout
			} else {
				cmd.Stdout = state.tty
			}

			cmd.Stderr = os.Stderr

			cmd.Run() // TODO handle errors
		} else {
			// no executable
			if state.done {
				os.Stdout.WriteString(line)
				os.Stdout.WriteString("\n")
			} else {
				state.tty.WriteString(line)
				state.tty.WriteString("\n")
			}
		}
	}

	return nil
}

type LineChangeState struct {
	shell       string
	tty         *os.File
	linebuffer  *[]rune
	show        bool
	inputOnly   bool
	execute     *string
	done        bool
	environment []string
}

// Start the interactive line-editor
func rl(show bool, inputOnly bool, execute *string) {
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
		show,
		inputOnly,
		execute,
		false,
		os.Environ(),
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
