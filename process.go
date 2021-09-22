package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/rivo/tview"
)

// Wait for started commands to complete.
func AwaitCommand(cmd *exec.Cmd, stdoutBuffer *bytes.Buffer, tui *TUI) {
	// wait performs cleanup tasks; without this a large number of threads pile-up in this process.

	cmd.Wait()

	tui.SetLineCount(stdoutBuffer)
	tui.UpdateScrollPosition()

	// TODO by default, scroll seems to lock to the bottom of the document. TODO may be annoying
	// if you scrolled in view mode and tried to apply highlighting / line-number respecting filters.
	tui.stdoutViewer.tview.ScrollToBeginning()
	tui.Draw()
}

type ClearWriter struct {
	view    *tview.TextView
	writer  io.Writer
	cleared bool
}

// Defer to writer, but clear text-view on first write just-before
// new output is written (minimising flickering in output)
func (tgt *ClearWriter) Write(data []byte) (n int, err error) {
	if !tgt.cleared {
		tgt.view.Lock()
		tgt.view.Clear()
		tgt.cleared = true
		tgt.view.Unlock()
	}

	return tgt.writer.Write(data)
}

func NewClearWriter(view *tview.TextView) *ClearWriter {
	return &ClearWriter{
		view,
		tview.ANSIWriter(view),
		false,
	}
}

// Given the user-input, and contextual information, start a provided command in the user's shell
// and point it at /dev/tty if in preview mode, or standard-output if the linebuffer is done. This command
// will have access to an environmental variable containing the user's input
func StartCommand(tui *TUI) (*exec.Cmd, error) {
	// run the provided command in the user's shell. We don't know for certain -c is the correct
	// flag, this wil vary between shells. but it works for zsh and bash.

	// only output the result of the last command-execution to standard-output; otherwise just show it on the tty
	done := tui.GetDone()
	ctx := tui.ctx
	lineBuffer := tui.state.lineBuffer

	cmd := exec.Command(ctx.shell, "-c", *ctx.execute)

	// is stdin present? If it is, StdinReader will have captured it.
	piped, err := StdinPiped()
	if err != nil {
		return cmd, err
	}

	if piped {
		// construct a new reader from stdin bytes
		cmd.Stdin = bytes.NewReader(ctx.stdin.Bytes())
	}

	varlist := []string{ENVAR_NAME_RL_INPUT + "=" + lineBuffer.content}

	for _, pair := range ctx.envVars {
		varlist = append(varlist, pair[0]+"="+pair[1])
	}

	// by default, go will use the current  process's environment. Merge RL_INPUT into that list and provide it to the command
	cmd.Env = append(ctx.environment, varlist...)

	var stdoutBuffer bytes.Buffer

	if done {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		outputView := NewClearWriter(tui.stdoutViewer.tview)

		// show intermixed, like a terminal would, by pipeing
		// everthing to outputview
		cmd.Stdout = io.MultiWriter(outputView, &stdoutBuffer)
		cmd.Stderr = outputView
	}
	// set the pgid so we can terminate this child-process and its descendents with one signal later, if we need to
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if done {
		// so something odd was happening here; before tui.Stop() was called all output was swallowed.
		// I imagine I screwed up with os.Stdout handling here.
		tui.Stop()

		fmt.Fprintf(os.Stderr, SubstitueCommand(ctx.execute, &lineBuffer.content)+"\n")
		finalErr := cmd.Run()

		return nil, finalErr
	} else {
		// start the command, but don't wait for the command to complete or error-check that it started

		cmd.Start()
		go AwaitCommand(cmd, &stdoutBuffer, tui)
	}

	if err != nil {
		return nil, err
	} else {
		return cmd, nil
	}
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
func (state LineChangeState) HandleUserUpdate(tui *TUI) (LineChangeState, error) {
	// no matter what we do, we don't need an old command still running; stop it
	state.StopProcess()

	done := tui.GetDone()

	if done && tui.ctx.inputOnly {
		// we don't case about final command execution; just print what
		// the user inputted and exit.
		tui.Stop()
		fmt.Println(tui.state.lineBuffer.content)

		go func(exitChan chan int) {
			exitChan <- 0
		}(tui.chans.exitCode)

		return state, nil
	}

	// call the command
	cmd, cmdErr := StartCommand(tui)

	// if done, handle exit codes
	if done {
		go func(exitChan chan int) {
			if exitError, ok := cmdErr.(*exec.ExitError); ok {
				exitChan <- exitError.ExitCode()
			} else if cmdErr != nil {
				// it faied, we don't know why
				exitChan <- 1
			} else {
				exitChan <- 0
			}
		}(tui.chans.exitCode)

		return state, nil
	}

	if cmdErr != nil {
		return state, cmdErr
	} else {
		state.cmd = cmd
	}

	return state, nil
}
