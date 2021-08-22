package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/eiannone/keyboard"
)

// Open /dev/tty with user write-only permissions. If it fails to open, return
// an error that will indicate this tool is being run in non-interactive mode
func OpenTTY() (*os.File, error) {
	fd, err := syscall.Open("/dev/tty", syscall.O_WRONLY, 00200)

	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "pipe"), nil
}

// Store variables that will changes as characters are received from the user
// and commands are executed.
type LineChangeState struct {
	lineBuffer *[]rune   // a pointer to an array of characters the user has entered into this application, excluding some special characters like backspaces.
	done       bool      // has a terminating character been received by the application?
	cmd        *exec.Cmd // a pointer to the command currently being executed, if rl is running in execute mode
}

type LineChangeCtx struct {
	shell       string   // the user's shell-variable
	tty         *os.File // a pointer to /dev/tty
	show        bool     // is the show option enabled? i.e should we avoid clearing the screen pre-command execution?
	inputOnly   bool     // should we only return the user's input (e.g lineBuffer) instead of the final command execution, if we're running in execute mode?
	execute     *string  // a string to execute in a user's shell
	environment []string // an array of this processes environmental variables
}

// Stop a running execute process by looking up the state's cmd variable,
// and if it's present send a SIGKILL signal to the child-process (the user's spawned shell) and
// the processes started by it. This is important to stop slow-running commands from making this tool
// feel laggy; we're running a process for the new user-input as fast as possible
func (state *LineChangeState) StopProcess() error {
	cmd := state.cmd

	if cmd == nil {
		return nil
	}

	pgid, err := syscall.Getpgid(cmd.Process.Pid)

	if err != nil {
		return err
	}

	// this seems like overkill (hah) but fzf sends this signal
	return syscall.Kill(-pgid, syscall.SIGKILL)
}

// an ANSI escape string to clear a screen (https://unix.stackexchange.com/questions/124762/how-does-clear-command-work)
const CLEAR_STRING = "\x1b\x5b\x48\x1b\x5b\x32\x4a"

// This command executes each time the user enters input, and may run attempt to run concurrently. It uses a
// mutex to avoid concurrency issues; and performs a few steps:
// - Stop all running child-processes
func OnUserInputChange(state LineChangeState, ctx *LineChangeCtx) (LineChangeState, error) {
	isExecuteMode := len(*ctx.execute) > 0

	if !ctx.show {
		ctx.tty.Write([]byte(CLEAR_STRING))
	}

	line := string(*state.lineBuffer)

	if !isExecuteMode {
		// no command to execute
		if state.done {
			os.Stdout.WriteString(line + "\n")
		} else {
			ctx.tty.WriteString(line + "\n")
		}
		return state, nil
	} else if state.done && ctx.inputOnly {
		// we're done, we only want the input line but not the command output
		os.Stdout.WriteString(line + "\n")

		return state, nil
	}

	state.StopProcess()

	cmd, startErr := StartCommand(state.done, line, ctx)

	if startErr != nil {
		return state, startErr
	} else {
		state.cmd = cmd
	}

	return state, nil
}

func StartCommand(done bool, line string, ctx *LineChangeCtx) (*exec.Cmd, error) {
	// run the provided command in the user's shell. We don't know for certain -c is the correct
	// flag, this wil vary between shells. but it works for zsh and bash.
	cmd := exec.Command(ctx.shell, "-c", *ctx.execute)

	// by default, go will use the current  process's environment. Merge RL_INPUT into that list and provide it to the command
	cmd.Env = append(ctx.environment, "RL_INPUT="+line)
	cmd.Stderr = os.Stderr

	// only output the result of the last command-execution to standard-output; otherwise just show it on the tty
	if done {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = ctx.tty
	}
	// set the pgid so we can terminate this child-process and its descendents with one signal later, if we need to
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// start the command, but don't wait for the command to complete or error-check that it started
	err := cmd.Start()

	go func(cmd *exec.Cmd) {
		// wait performs cleanup tasks; without this a large number of threads pile-up in this process.
		cmd.Wait()
	}(cmd)

	if err != nil {
		return nil, err
	} else {
		return cmd, nil
	}
}

// Given a character and a keypress code, return an updated user-input text-buffer and a boolean
// indicating whether a terminating character like Enter or Escape was received.
func UpdateLineBuffer(char rune, key keyboard.Key, lineBuffer []rune) ([]rune, bool) {
	if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
		if len(lineBuffer) == 0 {
			return []rune{}, false
		} else {
			return lineBuffer[:len(lineBuffer)-1], false
		}
	} else if key == keyboard.KeyEsc || key == keyboard.KeyEnter {
		return lineBuffer, true
	} else if key == keyboard.KeySpace {
		return append(lineBuffer, ' '), false
	}

	return append(lineBuffer, char), false
}

// Start the interactive line-editor with any provided CLI arguments
func RL(show bool, inputOnly bool, execute *string) int {

	if err := keyboard.Open(); err != nil {
		if strings.Contains(err.Error(), "/dev/tty") {
			fmt.Printf("RL: could not open /dev/tty. Are you running rl non-interactively?")
		} else {
			fmt.Printf("RL: failed to read from keyboard. %v\n", err)
		}
		// I hate this pattern but it honours deferred functions
		return 1
	}
	defer func() {
		keyboard.Close()
	}()

	tty, ttyErr := OpenTTY()

	if ttyErr != nil {
		fmt.Printf("RL: could not open /dev/tty. Are you running rl non-interactively?")
		return 1
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

	if ctx.shell == "" {
		fmt.Printf("RL: could not determine user's shell (e.g bash, zsh). Ensure $SHELL is set.")
		return 1
	}

	lineBuffer := []rune{}
	state := LineChangeState{
		&lineBuffer,
		false,
		nil,
	}

	var done bool
	for {
		// repeatedly read input from a keyboard, until some command
		// repeatedly get keys, until a terminating character like Escape or Enter is reached.
		char, key, err := keyboard.GetKey()

		// this library seems to mask keyboard signals, we need to
		// handle them ourselves. Using C style `raise` does not appear to be a good idea
		// https://github.com/golang/go/issues/19326
		if key == keyboard.KeyCtrlC || key == keyboard.KeyCtrlZ {
			return 0
		}

		if err != nil {
			fmt.Printf("RL: Keyboard read failed. %v\n", err)
			return 1
		}

		lineBuffer, done = UpdateLineBuffer(char, key, lineBuffer)

		state.lineBuffer = &lineBuffer
		state.done = done

		state, _ = OnUserInputChange(state, &ctx)

		if state.done {
			break
		}
	}

	return 0
}
