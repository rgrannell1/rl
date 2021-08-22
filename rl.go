package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/eiannone/keyboard"
)

const ENVAR_NAME_RL_INPUT = "RL_INPUT"

// Allow user writes, no other permissions
const USER_WRITE_OCTAL = 00200

// an ANSI escape string to clear a screen (https://unix.stackexchange.com/questions/124762/how-does-clear-command-work)
const CLEAR_STRING = "\x1b\x5b\x48\x1b\x5b\x32\x4a"

// Open /dev/tty with user write-only permissions. If it fails to open, return
// an error that will indicate this tool is being run in non-interactive mode
func OpenTTY() (*os.File, error) {
	fd, err := syscall.Open("/dev/tty", syscall.O_WRONLY, USER_WRITE_OCTAL)

	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), "pipe"), nil
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

	// this seems like overkill (hah) but fzf sends this signal rather than the SIGTERM I initially went with,
	// and I trust their decision
	return syscall.Kill(-pgid, syscall.SIGKILL)
}

// Takes the current application state, and some context variables, and run a few steps:
// - clear the terminal, if required
// - cleanup any old processes running
// - print the command-output, or input-text, to /dev/tty or standard output
func (state LineChangeState) HandleUserUpdate(ctx *LineChangeCtx) (LineChangeState, error) {
	isExecuteMode := len(*ctx.execute) > 0

	if !ctx.show {
		ctx.tty.Write([]byte(CLEAR_STRING))
	}

	if !isExecuteMode {
		// no command to execute
		line := state.lineBuffer.String()

		if state.lineBuffer.done {
			os.Stdout.WriteString(line + "\n")
		} else {
			ctx.tty.WriteString(line + "\n")
		}
		return state, nil
	} else if state.lineBuffer.done && ctx.inputOnly {
		// we're done, we only want the input line but not the command output
		os.Stdout.WriteString(state.lineBuffer.String() + "\n")

		return state, nil
	}

	state.StopProcess()

	cmd, startErr := StartCommand(state.lineBuffer, ctx)

	if startErr != nil {
		return state, startErr
	} else {
		state.cmd = cmd
	}

	return state, nil
}

// Given the user-input, and contextual information, start a provided command in the user's shell
// and point it at /dev/tty if in preview mode, or standard-output if the linebuffer is done. This command
// will have access to an environmental variable containing the user's input
func StartCommand(lineBuffer *LineBuffer, ctx *LineChangeCtx) (*exec.Cmd, error) {
	// run the provided command in the user's shell. We don't know for certain -c is the correct
	// flag, this wil vary between shells. but it works for zsh and bash.
	cmd := exec.Command(ctx.shell, "-c", *ctx.execute)

	// by default, go will use the current  process's environment. Merge RL_INPUT into that list and provide it to the command
	cmd.Env = append(ctx.environment, ENVAR_NAME_RL_INPUT+"="+lineBuffer.String())
	cmd.Stderr = os.Stderr

	// only output the result of the last command-execution to standard-output; otherwise just show it on the tty
	if lineBuffer.done {
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
func (lineBuffer LineBuffer) UpdateLine(char rune, key keyboard.Key) LineBuffer {
	if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
		return lineBuffer.Backspace()
	} else if key == keyboard.KeyEsc || key == keyboard.KeyEnter {
		lineBuffer.done = true
		return lineBuffer
	} else if key == keyboard.KeySpace {
		return lineBuffer.AddChar(' ')
	}

	return lineBuffer.AddChar(char)
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

	lineBuffer := LineBuffer{}
	state := LineChangeState{&lineBuffer, nil}

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

		lineBuffer = lineBuffer.UpdateLine(char, key)
		state.lineBuffer = &lineBuffer

		state, _ = state.HandleUserUpdate(&ctx)

		if state.lineBuffer.done {
			break
		}
	}

	return 0
}
