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

// Open /dev/tty with user write permissions.
func OpenTTY() (*os.File, error) {
	fd, err := syscall.Open("/dev/tty", syscall.O_WRONLY, 00200)

	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "pipe"), nil
}

type LineChangeState struct {
	lineBuffer *[]rune
	done       bool
	cmd        *exec.Cmd
}

type LineChangeCtx struct {
	shell       string
	tty         *os.File
	show        bool
	inputOnly   bool
	execute     *string
	environment []string
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

const CLEAR_STRING = "\033[H\033[2J"

// Run each time a user inputs a character
func OnUserInputChange(stateChan chan LineChangeState, cmdLock *sync.Mutex, ctx *LineChangeCtx) {
	state := <-stateChan

	cmdLock.Lock()
	defer func() {
		cmdLock.Unlock()
	}()

	StopProcess(&state)

	if !ctx.show {
		ctx.tty.Write([]byte(CLEAR_STRING))
	}

	line := string(*state.lineBuffer)

	if len(*ctx.execute) == 0 {
		// no command to execute
		if state.done {
			os.Stdout.WriteString(line + "\n")
		} else {
			ctx.tty.WriteString(line + "\n")
		}
		return
	} else if state.done && ctx.inputOnly {
		// we're done, we only want the input line but not the command output
		os.Stdout.WriteString(line + "\n")

		stateChan <- state
		return
	}

	// run the provided command in the users shell
	cmd := exec.Command(ctx.shell, "-c", *ctx.execute)

	// by default, go will use the current  process's environment. Add RL_INPUT to the list.
	cmd.Env = append(ctx.environment, "RL_INPUT="+line)
	cmd.Stderr = os.Stderr

	// only output the last command to standard-output by default; otherwise just show it on the tty
	if state.done {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = ctx.tty
	}

	state.cmd = cmd

	// non-blocking command-start
	state.cmd.Start()

	stateChan <- state
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

	tty, ttyErr := OpenTTY()

	if ttyErr != nil {
		fmt.Printf("RL: could not open /dev/tty. Are you running rl non-interactively?")
		os.Exit(1)
	}
	defer func() {
		tty.Close()
	}()

	ctx := LineChangeCtx{
		os.Getenv("SHELL"),
		tty,
		show,
		inputOnly,
		execute,
		os.Environ(),
	}

	stateChan := make(chan LineChangeState)
	defer func() {
		close(stateChan)
	}()

	doneChan := make(chan bool)
	defer func() {
		close(doneChan)
	}()

	cmdLock := &sync.Mutex{}

	lineBuffer := []rune{}
	state := LineChangeState{
		&lineBuffer,
		false,
		nil,
	}

	var done bool
	for {
		// repeatedly get keys, until a terminating character is reached
		char, key, err := keyboard.GetKey()

		if err != nil {
			fmt.Printf("RL: Keyboard read failed. %v\n", err)
			os.Exit(1)
		}

		lineBuffer, done = UpdateLineBuffer(char, key, lineBuffer)

		state.lineBuffer = &lineBuffer
		state.done = done

		go OnUserInputChange(stateChan, cmdLock, &ctx)
		stateChan <- state

		state = <-stateChan

		if state.done {
			break
		}
	}
}

// Update the user input line-buffer
func UpdateLineBuffer(char rune, key keyboard.Key, lineBuffer []rune) ([]rune, bool) {
	if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
		if len(lineBuffer) == 0 {
			return []rune{}, false
		} else {
			return lineBuffer[:len(lineBuffer)-1], false
		}
	} else if key == keyboard.KeyCtrlC || key == keyboard.KeyEsc || key == keyboard.KeyEnter {
		return lineBuffer, true
	} else if key == keyboard.KeySpace {
		return append(lineBuffer, ' '), false
	}

	return append(lineBuffer, char), false
}
