package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/eiannone/keyboard"
)

// Open /dev/tty
func OpenTTY() (*os.File, error) {
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
}

// Stop a running execute process
func StopProcess(state *LineChangeState) {
	if state.cmd != nil {
		if state.cmd.Process != nil {
			state.cmd.Process.Signal(syscall.SIGINT)
			state.cmd = nil
		}
	}
}

// Run each time a user inputs a character
func OnUserInputChange(stateChan chan LineChangeState, cmdLock *sync.Mutex) {
	state := <-stateChan

	cmdLock.Lock()
	StopProcess(&state)

	if !state.show {
		state.tty.Write([]byte("\033[H\033[2J"))
	}

	line := string(*state.linebuffer)

	if len(*state.execute) == 0 {
		// no command to execute
		if state.done {
			os.Stdout.WriteString(line + "\n")
		} else {
			state.tty.WriteString(line + "\n")
		}
		return
	} else if state.done && state.inputOnly {
		// we're done, we only want the input line but not the command output
		os.Stdout.WriteString(line + "\n")

		stateChan <- state
		return
	}

	// run the provided command in the users shell
	cmd := exec.Command(state.shell, "-c", *state.execute)

	// by default, go will use the current  process's environment. Add RL_INPUT to the list.
	cmd.Env = append(state.environment, "RL_INPUT="+line)
	cmd.Stderr = os.Stderr

	// only output the last command to standard-output by default; otherwise just show it on the tty
	if state.done {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = state.tty
	}

	state.cmd = cmd

	// non-blocking command-start;
	state.cmd.Start()

	stateChan <- state

	cmdLock.Unlock()
}

// Start the interactive line-editor
func rl(show bool, inputOnly bool, execute *string) {

	if err := keyboard.Open(); err != nil {
		if strings.Contains(err.Error(), "/dev/tty") {
			fmt.Printf("RL: could not open /dev/tty. Are you running rl non-interactively?")
		} else {
			fmt.Printf("RL: failed to read from keyboard. %v\n", err)
		}
		os.Exit(1)
	}
	defer func() {
		keyboard.Close()
	}()

	linebuffer := []rune{}

	tty, ttyErr := OpenTTY()

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
		go OnUserInputChange(stateChan, cmdLock)

		stateChan <- state
		state = <-stateChan

		if state.done {
			break
		}
	}
}
