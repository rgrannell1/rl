package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
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

type LineChangeState struct {
	shell       string
	tty         *os.File
	linebuffer  *[]rune
	show        bool
	inputOnly   bool
	execute     *string
	done        bool
	environment []string
	cmd         *exec.Cmd
	running     bool
}

func onLinebufferChange(stateChan chan LineChangeState, cmdLock *sync.Mutex) {
	state := <-stateChan

	// LOCKED
	cmdLock.Lock()
	// first stop the existing process

	if state.cmd != nil {
		if state.cmd.Process != nil {
			state.cmd.Process.Signal(syscall.SIGINT)
			state.cmd = nil
		}
	}

	if !state.show {
		// we don't really care enough about clear errors to spam the console; ignore the errors,
		// and attempt to clear the TTY. This program should always be run in interactive mode.

		state.tty.Write([]byte("\033[H\033[2J"))
	}

	line := string(*state.linebuffer)

	if len(*state.execute) == 0 {
		// no executable
		if state.done {
			os.Stdout.WriteString(line)
			os.Stdout.WriteString("\n")
		} else {
			state.tty.WriteString(line)
			state.tty.WriteString("\n")
		}
	} else {
		if state.done && state.inputOnly {
			// we're done, we only want the input line but not the command output
			os.Stdout.WriteString(line)
			os.Stdout.WriteString("\n")

			stateChan <- state
			return
		}

		// run the provided command in the users shell
		//toExec := fmt.Sprintf("export RL_INPUT=\"%s*\"; %s", line, *state.execute)
		toExec := fmt.Sprintf(*state.execute)
		state.cmd = exec.Command(state.shell, "-c", toExec)

		// by default, go will use the current  process's environment. Add RL_INPUT to the list.
		state.cmd.Env = append(state.environment, "RL_INPUT="+line)

		// only output the last command to standard-output by default
		if state.done {
			state.cmd.Stdout = os.Stdout
		} else {
			state.cmd.Stdout = state.tty
		}

		state.cmd.Stderr = os.Stderr

		// this will not block
		state.cmd.Start()

		stateChan <- state

		cmdLock.Unlock()
	}
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
		nil,
		false,
	}

	stateChan := make(chan LineChangeState)
	defer func() {
		close(stateChan)
	}()

	doneChan := make(chan bool)
	defer func() {
		close(doneChan)
	}()

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

		cmdLock := &sync.Mutex{}

		// handle command execution!! Factor this out when I can
		go onLinebufferChange(stateChan, cmdLock)

		stateChan <- state
		state = <-stateChan

		if state.done {
			break
		}
	}
}
