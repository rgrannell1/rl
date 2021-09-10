package main

import (
	"os"
	"os/exec"
	"syscall"
)

// Given the user-input, and contextual information, start a provided command in the user's shell
// and point it at /dev/tty if in preview mode, or standard-output if the linebuffer is done. This command
// will have access to an environmental variable containing the user's input
func StartCommand(lineBuffer *LineBuffer, ctx *LineChangeCtx) (*exec.Cmd, error) {
	// run the provided command in the user's shell. We don't know for certain -c is the correct
	// flag, this wil vary between shells. but it works for zsh and bash.
	cmd := exec.Command(ctx.shell, "-c", *ctx.execute)

	// by default, go will use the current  process's environment. Merge RL_INPUT into that list and provide it to the command
	cmd.Env = append(ctx.environment, ENVAR_NAME_RL_INPUT+"="+lineBuffer.content)

	// only output the result of the last command-execution to standard-output; otherwise just show it on the tty
	if lineBuffer.done {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		nastyGlobalState = true
		cmd.Stdout = ctx
		cmd.Stderr = ctx // this could be refined
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

var nastyGlobalState = true

// Implement IO.Writer for Ctx so we can clear _just before_ the new command text is received,
// so we don't see flashes and latency
func (ctx *LineChangeCtx) Write(data []byte) (n int, err error) {
	// this will panic if a lock isn't set!
	if nastyGlobalState {
		ctx.tgt.Lock()
		ctx.tgt.Clear()
		nastyGlobalState = false
		ctx.tgt.Unlock()
	}

	return ctx.tgt.Write(data)
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
// - print the command-output to tview or standard output
func (state LineChangeState) HandleUserUpdate(ctx *LineChangeCtx) (LineChangeState, error) {
	state.StopProcess()

	cmd, startErr := StartCommand(state.lineBuffer, ctx)

	if startErr != nil {
		return state, startErr
	} else {
		state.cmd = cmd
	}

	return state, nil
}