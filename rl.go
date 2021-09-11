package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/smallnest/ringbuffer"
)

const ENVAR_NAME_RL_INPUT = "RL_INPUT"

// Update the UI header
func updateHeader(header *tview.TextView, command string, buffer *LineBuffer) {
	summary := strings.ReplaceAll(command, "$"+ENVAR_NAME_RL_INPUT, "[red]"+buffer.content+"[default]")

	header.SetText("rl: " + summary)
}

const PROMPT = "> "

// Start the interactive line-editor with any provided CLI arguments
func RL(inputOnly bool, execute *string) int {
	tty, ttyErr := OpenTTY()
	cfg, cfgErr := InitConfig()

	if cfgErr != nil {
		fmt.Printf("RL: Failed to read configuration: %s\n", cfgErr)
		return 1
	}

	if ttyErr != nil {
		fmt.Printf("RL: could not open /dev/tty. Are you running rl non-interactively?")
		return 1
	}
	tty.Close()

	piped, pipeErr := StdinPiped()
	if pipeErr != nil {
		fmt.Printf("RL: could not inspect whether sdin was piped in.")

		return 1
	}

	stdin := ringbuffer.New(1000 * 10) // 10MB

	ctx := LineChangeCtx{
		os.Getenv("SHELL"),
		inputOnly,
		execute,
		os.Environ(),
		nil,
		stdin,
	}

	if ctx.shell == "" {
		fmt.Printf("RL: could not determine user's shell (e.g bash, zsh). Ensure $SHELL is set.")
		return 1
	}

	// read from standard input and redirect to subcommands. Input can be infinite,
	// so manage this read from a goroutine an read into a circular buffer
	if piped {
		go StdinReader(stdin)
	}

	histChan := make(chan *History)
	defer func() {
		close(histChan)
	}()

	if cfg.Config.SaveHistory {
		go HistoryWriter(histChan, cfg)
	}

	linebuffer := LineBuffer{}
	state := LineChangeState{
		lineBuffer: &linebuffer,
		cmd:        nil,
	}

	app := tview.NewApplication()
	// -- we actually want to do the opposite; this prevents scroll from breaking things.
	app.EnableMouse(true)
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		return nil, 0
	})

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault

	header := tview.NewTextView().
		SetText("rl: " + *execute).SetTextColor(tcell.ColorDefault).
		SetDynamicColors(true)

	main := tview.NewTextView().
		SetText("").
		SetDynamicColors(true).
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

			if cfg.Config.SaveHistory {
				histChan <- &History{
					Input:    text,
					Command:  strings.ReplaceAll(*execute, "$"+ENVAR_NAME_RL_INPUT, text),
					Template: *execute,
					Time:     time.Now(),
				}
			}

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
