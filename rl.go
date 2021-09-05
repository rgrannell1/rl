package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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

// Given the user-input, and contextual information, start a provided command in the user's shell
// and point it at /dev/tty if in preview mode, or standard-output if the linebuffer is done. This command
// will have access to an environmental variable containing the user's input
func StartCommand(lineBuffer *LineBuffer, ctx *LineChangeCtx) (*exec.Cmd, error) {
	// run the provided command in the user's shell. We don't know for certain -c is the correct
	// flag, this wil vary between shells. but it works for zsh and bash.
	cmd := exec.Command(ctx.shell, "-c", *ctx.execute)

	// by default, go will use the current  process's environment. Merge RL_INPUT into that list and provide it to the command
	cmd.Env = append(ctx.environment, ENVAR_NAME_RL_INPUT+"="+lineBuffer.content)
	cmd.Stderr = os.Stderr

	// only output the result of the last command-execution to standard-output; otherwise just show it on the tty
	if lineBuffer.done {
		cmd.Stdout = os.Stdout
	} else {
		nastyGlobalState = true
		cmd.Stdout = ctx
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

func updateHeader(header *tview.TextView, command string, buffer *LineBuffer) {
	summary := strings.ReplaceAll(command, "$"+ENVAR_NAME_RL_INPUT, "[red]"+buffer.content+"[default]")

	header.SetText("rl: " + summary)
}

const PROMPT = "> "

// Start the interactive line-editor with any provided CLI arguments
func RL(inputOnly bool, execute *string) int {
	tty, ttyErr := OpenTTY()

	if ttyErr != nil {
		fmt.Printf("RL: could not open /dev/tty. Are you running rl non-interactively?")
		return 1
	}
	tty.Close()

	ctx := LineChangeCtx{
		os.Getenv("SHELL"),
		inputOnly,
		execute,
		os.Environ(),
		nil,
	}

	if ctx.shell == "" {
		fmt.Printf("RL: could not determine user's shell (e.g bash, zsh). Ensure $SHELL is set.")
		return 1
	}

	linebuffer := LineBuffer{}
	state := LineChangeState{
		lineBuffer: &linebuffer,
		cmd:        nil,
	}

	app := tview.NewApplication()
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault

	header := tview.NewTextView().
		SetText("rl: " + *execute).SetTextColor(tcell.ColorDefault).
		SetDynamicColors(true)

	main := tview.NewTextView().
		SetText("").
		SetTextColor(tcell.ColorDefault).
		SetChangedFunc(func() {
			app.Draw()
		})

	ctx.tgt = main

	rlInput := tview.NewInputField()
	rlInput.
		SetChangedFunc(func(text string) {
			state.lineBuffer.content = text
			state, _ = state.HandleUserUpdate(&ctx)

			updateHeader(header, *execute, state.lineBuffer)
		}).
		SetLabel(PROMPT).
		SetDoneFunc(func(key tcell.Key) {
			state.lineBuffer.done = true

			app.Stop() // exits on arrow
			state, _ = state.HandleUserUpdate(&ctx)
		})

	grid := tview.NewGrid().
		SetRows(2, 0, 1).
		SetColumns(30, 0, 30).SetBorders(false).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false).
		AddItem(rlInput, 2, 0, 1, 3, 1, 0, true)

	app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		return e
	})

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		fmt.Printf("RL: Application crashed! %v", err)
		return 1
	}

	return 0
}
