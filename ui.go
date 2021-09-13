package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Update the UI header
func updateHeader(header *tview.TextView, command string, buffer *LineBuffer) {
	summary := strings.ReplaceAll(command, "$"+ENVAR_NAME_RL_INPUT, "[red]"+buffer.content+"[default]")

	header.SetText("rl: " + summary)
}

func CreateHeader(execute *string) *tview.TextView {
	return tview.NewTextView().
		SetText("rl: " + *execute).SetTextColor(tcell.ColorDefault).
		SetDynamicColors(true)
}

func CreateApp() *tview.Application {
	// -- declare Tview application --
	app := tview.NewApplication()
	// -- we actually want to do the opposite; this prevents mousewheel scroll from breaking things.
	app.EnableMouse(true)
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		return nil, 0
	})

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault

	return app
}

func CreateTextView(app *tview.Application) *tview.TextView {
	return tview.NewTextView().
		SetText("").
		SetDynamicColors(true).
		SetTextColor(tcell.ColorDefault).
		SetChangedFunc(func() {
			app.Draw()
		})
}

func (ctx *LineChangeCtx) CreateUI(state LineChangeState, cfg ConfigOpts, histChan chan *History) int {
	execute := ctx.execute
	app := CreateApp()
	header := CreateHeader(execute)
	stdoutViewer := CreateTextView(app)

	ctx.tgt = stdoutViewer

	rlInput := tview.NewInputField()
	rlInput.
		SetChangedFunc(func(text string) {
			state.lineBuffer.content = text
			state, _ = state.HandleUserUpdate(ctx)

			if cfg.Config.SaveHistory {
				histChan <- &History{
					Input:    text,
					Command:  SubstitueCommand(execute, &text),
					Template: *execute,
					Time:     time.Now(),
				}
			}

			updateHeader(header, *execute, state.lineBuffer)
		}).
		SetLabel(PROMPT).
		SetDoneFunc(func(key tcell.Key) {
			// this is invoked for KeyEnter, KeyEscape, KeyTab, KeyDown, KeyUp, KeyBacktab.

			if key == tcell.KeyUp || key == tcell.KeyDown {
				app.SetFocus(stdoutViewer)
			} else {
				app.Stop() // exits on arrow
			}

			state, _ = state.HandleUserUpdate(ctx)
		})

	grid := tview.NewGrid().
		SetRows(2, 0, 1).
		SetColumns(30, 0, 30).SetBorders(false).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(stdoutViewer, 1, 0, 1, 3, 0, 0, false).
		AddItem(rlInput, 2, 0, 1, 3, 1, 0, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		fmt.Printf("RL: Application crashed! %v", err)
		return 1
	}

	return 0
}
