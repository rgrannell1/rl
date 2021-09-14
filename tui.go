package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RLs UI app
type TUI struct {
	state          *LineChangeState
	cfg            *ConfigOpts
	ctx            *LineChangeCtx
	app            *TUIApp
	commandPreview *TUICommandPreview
	linePosition   *TUILinePosition
	stdoutViewer   *TUIStdoutView
	commandInput   *TUICommandInput
	chans          struct {
		history chan *History
	}
	mode PromptMode
}

// Update the line-position element based on the current
// scroll-position
func (tui *TUI) UpdateScrollPosition() {
	stdout := tui.stdoutViewer.tview

	row, col := stdout.GetScrollOffset()
	_, _, _, height := stdout.GetInnerRect()

	tui.linePosition.row = row
	tui.linePosition.col = col
	tui.linePosition.height = height
	lineCount := tui.linePosition.lineCount

	endRow := row + height - 1

	rowStr := fmt.Sprint(row + 1)         // lines are normally one-indexed
	endRowStr := fmt.Sprint(endRow)       // the last line shown in the buffer
	lineCountStr := fmt.Sprint(lineCount) // the total line count produced by standard-output last execution

	var percentStr = ""

	if lineCount == 0 {
		percentStr = ""
	} else {
		ratio := float64(endRow) / float64(lineCount)
		percentStr = fmt.Sprint(math.Round(1_000.0*ratio)/10.0) + "%"
	}

	tui.linePosition.tview.SetText("line " + rowStr + "-" + endRowStr + "/" + lineCountStr + "    [blue]" + percentStr + "[blue]")
}

// Invert text command-input
func (tui *TUI) InvertCommandInput() {
	//input := tui.commandInput

	// TODO pick up system colours
	//input.tview.SetFieldBackgroundColor()
	//input.tview.SetFieldTextColor()
}

// Set initial theme overrides, so tview uses default
// system colours rather than tcell theme overrides
func (tui *TUI) SetTheme() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault
}

// Redraw the application
func (tui *TUI) Draw() {
	tui.app.tview.Draw()
}

// The application subcomponent, and operations on them
type TUIApp struct {
	tview *tview.Application
}

// Focus on stdout viewer
func (tui *TUI) SetStdoutViewerFocus() {
	tui.app.tview.SetFocus(tui.stdoutViewer.tview)
	tui.SetMode(ViewPrompt)

	// don't show command preview
	tui.commandInput.tview.SetText("")
}

// Focus on input
func (tui *TUI) SetInputFocus() {
	tui.app.tview.SetFocus(tui.commandInput.tview)
	tui.SetMode(CommandPrompt)
}

// Store RL's TUI
func (tui *TUI) Stop() {
	tui.app.tview.Stop() // exits on arrow
}

// this is not very readable; here are the AddItem definitions
// (p tview.Primitive, row int, column int, rowSpan int, colSpan int, minGridHeight int, minGridWidth int, focus bool) *tview.Grid
func (tui *TUI) Grid() *tview.Grid {
	return tview.NewGrid().
		SetRows(2, 0, 2).
		SetColumns(30, 0, 30).SetBorders(false).
		AddItem(tui.commandPreview.tview, 0, 0, 1, 2, 0, 0, false).
		AddItem(tui.linePosition.tview, 0, 2, 1, 1, 0, 0, false).
		AddItem(tui.stdoutViewer.tview, 1, 0, 1, 3, 0, 0, false).
		AddItem(tui.commandInput.tview, 2, 0, 1, 3, 1, 0, true)
}

// Start RL's TUI, and handle failures
func (tui *TUI) Start() int {
	grid := tui.Grid()

	// start the tview application
	if err := tui.app.tview.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		fmt.Printf("RL: Application crashed! %v", err)
		return 1
	}

	return 0
}

//  The preview element showing a preview of the command that will be executed
type TUICommandPreview struct {
	tview *tview.TextView
}

// Update the UI header based on user input
func (prev *TUICommandPreview) UpdateText(command string, buffer *LineBuffer) {
	summary := strings.ReplaceAll(command, "$"+ENVAR_NAME_RL_INPUT, "[red]"+buffer.content+"[default]")
	prev.tview.SetText("rl: " + summary)
}

// A component for the line-position in the stdout viewer
type TUILinePosition struct {
	tview     *tview.TextView
	lineCount int
	height    int
	row       int
	col       int
}

// A component for the stdout viewer
type TUIStdoutView struct {
	tview *tview.TextView
}

// A component for the RL text-input field
type TUICommandInput struct {
	tview *tview.InputField
}

// Set prompt
func (tui *TUI) SetMode(mode PromptMode) {
	tui.mode = mode

	if mode == CommandPrompt {
		tui.commandInput.tview.SetLabel(PROMPT_CMD)
	} else if mode == ViewPrompt {
		tui.commandInput.tview.SetLabel(PROMPT_VIEW)
	} else if mode == HelpPrompt {
		tui.commandInput.tview.SetLabel(HELP_VIEW)
	}
}

// Create the command-preview element; this will show what the user is actually executing
func NewCommandPreview(execute *string) *TUICommandPreview {
	part := tview.NewTextView().
		SetText("rl: " + *execute).SetTextColor(tcell.ColorDefault).
		SetDynamicColors(true)

	return &TUICommandPreview{part}
}

// Create a header widget that shows the current scroll position in
// the standard output viewer.
func NewLinePosition() *TUILinePosition {
	part := tview.NewTextView().
		SetText("").
		SetTextColor(tcell.ColorDefault).
		SetDynamicColors(true)

	return &TUILinePosition{part, 0, 0, 1, 0}
}

// Create RL tview application
func NewRLApp() *TUIApp {
	// -- declare Tview application --
	app := tview.NewApplication()
	// -- we actually want to do the opposite; this prevents mousewheel scroll from breaking things.
	app.EnableMouse(true)
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		return nil, 0
	})

	// if key-event filtering is needed, it can be applied here
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})

	return &TUIApp{app}
}

func NewStdoutView(tui *TUI) *TUIStdoutView {
	part := tview.NewTextView().
		SetText("").
		SetDynamicColors(true).
		SetTextColor(tcell.ColorDefault)

	part.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		key := event.Key()
		ch := event.Rune()

		if ch == '?' {
			tui.SetMode(HelpPrompt)
		} else if ch == '/' {
			tui.SetInputFocus()
			return nil
		} else if key == tcell.KeyUp {
			tui.UpdateScrollPosition()
		} else if key == tcell.KeyDown {
			tui.UpdateScrollPosition()
		} else if key == tcell.KeyEsc {
			tui.Stop()
		}

		return event
	})

	// this breaks things?
	//part.Focus(func(self tview.Primitive) {
	//	tui.commandInput.tview.SetBackgroundColor(tcell.ColorDefault)
	//	tui.commandInput.tview.SetFieldTextColor(tcell.ColorDefault)
	//})

	return &TUIStdoutView{part}
}

func NewCommandInput(tui *TUI) *TUICommandInput {
	ctx := tui.ctx
	cfg := tui.cfg
	state := *tui.state
	execute := ctx.execute

	commandInput := tview.NewInputField()

	commandInput.
		SetChangedFunc(func(text string) {
			state.lineBuffer.content = text
			state, _ = state.HandleUserUpdate(tui)

			if cfg.Config.SaveHistory {
				tui.chans.history <- &History{
					Input:    text,
					Command:  SubstitueCommand(execute, &text),
					Template: *execute,
					Time:     time.Now(),
				}
			}

			tui.commandPreview.UpdateText(*execute, state.lineBuffer)
		}).
		SetLabel(PROMPT_CMD).
		SetDoneFunc(func(key tcell.Key) {
			// this is invoked for KeyEnter, KeyEscape, KeyTab, KeyDown, KeyUp, KeyBacktab.

			if key == tcell.KeyUp {
				tui.SetStdoutViewerFocus()
				tui.UpdateScrollPosition()
			} else if key == tcell.KeyDown {
				tui.SetStdoutViewerFocus()
				tui.UpdateScrollPosition()
			} else if key == tcell.KeyEsc {
				tui.SetStdoutViewerFocus()
				tui.SetMode(ViewPrompt)
			} else {
				tui.Stop()
			}

			state, _ = state.HandleUserUpdate(tui)
		}).
		Focus(func(self tview.Primitive) {
			tui.InvertCommandInput()
		})

	return &TUICommandInput{commandInput}
}

func NewUI(state LineChangeState, cfg *ConfigOpts, ctx *LineChangeCtx, histChan chan *History) *TUI {
	execute := ctx.execute

	tui := TUI{}
	tui.mode = CommandPrompt
	tui.state = &state
	tui.cfg = cfg
	tui.ctx = ctx

	tui.SetTheme()
	tui.chans.history = histChan

	tui.app = NewRLApp()
	tui.commandPreview = NewCommandPreview(execute)
	tui.linePosition = NewLinePosition()
	tui.stdoutViewer = NewStdoutView(&tui)
	tui.commandInput = NewCommandInput(&tui)

	ctx.tgt = tui.stdoutViewer.tview // TODO bad, weird

	tui.InvertCommandInput()

	return &tui
}
