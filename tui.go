package main

import (
	"bytes"
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
	latency        *TUILatencyViewer
	linePosition   *TUILinePosition
	stdoutViewer   *TUITextViewer
	commandInput   *TUICommandInput
	helpBar        *TUIHelpBar
	chans          struct {
		history  chan *History
		exitCode chan int
	}
	mode      PromptMode
	textAlign int
	history   HistoryCursor
}

// provide some display of how long slow commands ran for
func (tui *TUI) UpdateRuntime(diff time.Duration) {
	ms := diff.Milliseconds()
	msg := fmt.Sprint(ms) + "ms"

	if ms < 100 {
		msg = "[green]" + msg + "[-:-:-]"
	} else if ms < 300 {
		msg = "[yellow]" + msg + "[-:-:-]"
	} else {
		msg = "[red]" + msg + "[-:-:-]"
	}

	tui.latency.tview.SetText(msg)
	tui.Draw()
}

// Update the line-position element based on the current
// scroll-position
func (tui *TUI) UpdateScrollPosition() {
	stdout := tui.stdoutViewer.tview
	stdout.Lock()

	row, col := stdout.GetScrollOffset()
	_, _, _, height := stdout.GetInnerRect()
	stdout.Unlock()

	tui.linePosition.row = row
	tui.linePosition.col = col
	tui.linePosition.height = height
	lineCount := tui.linePosition.lineCount

	endRow := math.Min(float64(row+height-1), float64(lineCount))

	rowStr := fmt.Sprint(row + 1)         // lines are normally one-indexed
	endRowStr := fmt.Sprint(endRow)       // the last line shown in the buffer
	lineCountStr := fmt.Sprint(lineCount) // the total line count produced by standard-output last execution

	var percentStr = ""

	if lineCount == 0 {
		percentStr = ""
	} else {
		ratio := math.Min(math.Max(0, float64(endRow)/float64(lineCount)), 1)

		percentStr = fmt.Sprint(math.Round(1_000.0*ratio)/10.0) + "%"
	}

	tui.linePosition.tview.SetText("line " + rowStr + "-" + endRowStr + " / " + lineCountStr + "    [blue]" + percentStr + "[blue]")
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

// Update the line-count based on stdout
func (tui *TUI) SetLineCount(stdoutBuffer *bytes.Buffer) {
	count := LineCounter(stdoutBuffer) // TODO this does not work reliably
	tui.linePosition.lineCount = count

	// clear if empty
	if tui.linePosition.lineCount == 0 {
		tui.stdoutViewer.tview.SetText("")
	}
}

func (tui *TUI) GetDone() bool {
	return tui.state.lineBuffer.done
}

// The application subcomponent, and operations on them
type TUIApp struct {
	tview *tview.Application
}

// Focus on stdout viewer
func (tui *TUI) SetStdoutViewerFocus() {
	tui.app.tview.SetFocus(tui.stdoutViewer.tview)

	// show a blue label, to make it obvious we switched mode
	tui.commandInput.tview.SetLabelColor(tcell.ColorBlue)
}

// Focus on input
func (tui *TUI) SetInputFocus() {
	tui.app.tview.SetFocus(tui.commandInput.tview)
	tui.commandInput.tview.SetLabelColor(tcell.ColorRed)
}

func (tui *TUI) ScrollHistoryBack() {

}

func (tui *TUI) ScrollHistoryForward() {

}

// Store RL's TUI
func (tui *TUI) Stop() {
	tui.app.tview.Stop() // exits on arrow
}

// this is not very readable; here are the AddItem definitions
// (p tview.Primitive, row int, column int, rowSpan int, colSpan int, minGridHeight int, minGridWidth int, focus bool) *tview.Grid
func (tui *TUI) Grid() *tview.Grid {
	return tview.NewGrid().
		SetRows(COMMAND_AND_LINE_ROWS, STDOUT_ROWS, SPACE_ROWS, HELP_ROWS, COMMAND_ROWS).
		SetColumns(-14, -6, -1).SetBorders(false).
		// add each item in a grid
		AddItem(tui.commandPreview.tview, ROW_0, COL_0, ROWSPAN_1, COLSPAN_1, MINWIDTH_0, MINHEIGHT_0, DONT_FOCUS).
		AddItem(tui.linePosition.tview, ROW_0, COL_1, ROWSPAN_1, COLSPAN_1, MINWIDTH_0, MINHEIGHT_0, DONT_FOCUS).
		AddItem(tui.latency.tview, ROW_0, COL_2, ROWSPAN_1, COLSPAN_1, MINWIDTH_0, MINHEIGHT_0, DONT_FOCUS).
		AddItem(tui.stdoutViewer.tview, ROW_1, COL_0, ROWSPAN_1, COLSPAN_3, MINWIDTH_0, MINHEIGHT_0, DONT_FOCUS).
		AddItem(tview.NewTextView(), ROW_2, COL_0, ROWSPAN_1, COLSPAN_3, MINWIDTH_1, MINHEIGHT_0, DONT_FOCUS).
		AddItem(tui.helpBar.tview, ROW_3, COL_0, ROWSPAN_1, COLSPAN_3, MINWIDTH_1, MINHEIGHT_0, DONT_FOCUS).
		AddItem(tui.commandInput.tview, ROW_4, COL_0, ROWSPAN_1, COLSPAN_3, MINWIDTH_0, MINHEIGHT_0, FOCUS)
}

// Start RL's TUI, and handle failures
func (tui *TUI) Start() int {
	defer close(tui.chans.exitCode)

	grid := tui.Grid()

	// start the tview application
	if err := tui.app.tview.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		fmt.Printf("RL: Application crashed! %v", err)
		return 1
	}

	return <-tui.chans.exitCode
}

//  The preview element showing a preview of the command that will be executed
type TUICommandPreview struct {
	tview *tview.TextView
}

type TUILatencyViewer struct {
	tview *tview.TextView
}

// Update the UI header based on user input
func (prev *TUICommandPreview) UpdateText(command string, buffer *LineBuffer, envVars *[][]string) {

	summary := strings.ReplaceAll(command, "$"+ENVAR_NAME_RL_INPUT, "[red]"+buffer.content+"[default]")

	for _, pair := range *envVars {
		varName := "$" + pair[0]
		// it might be nice to show env-vars, but these can contain passwords. This is a saner default.
		highlight := "[blue]" + varName + "[default]"

		summary = strings.ReplaceAll(summary, varName, highlight)
	}

	prev.tview.SetText("rl: " + "[::r]" + summary + "[-:-:-]")
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
type TUITextViewer struct {
	tview       *tview.TextView
	withDefault bool
}

// A component for the RL text-input field
type TUICommandInput struct {
	tview *tview.InputField
}

type TUIHelpBar struct {
	tview *tview.TextView
}

// Set prompt
func (tui *TUI) SetMode(mode PromptMode) {
	currMode := tui.mode
	tui.mode = mode

	if mode == EditMode {
		// EditMode switches
		tui.helpBar.tview.SetText(HELP_EDIT)
		tui.commandInput.tview.SetLabel(PROMPT_EDIT)
		tui.SetInputFocus()
	} else if mode == ViewMode {
		// Viewmode switches

		tui.helpBar.tview.SetText(HELP_VIEW)
		tui.commandInput.tview.SetLabel(PROMPT_VIEW)
		tui.SetStdoutViewerFocus()

		if currMode == HelpMode {
			// update the line-count in the buffer after switching
			tui.UpdateScrollPosition()
		}
	} else if mode == HelpMode {
		// Helpmode switches

		// TODO await lock to prevent clash with slow-writing command. Or just kill command
		tui.state.StopProcess()

		// TODO update line-count

		tui.helpBar.tview.SetText(HELP_HELP)
		tui.commandPreview.tview.SetText("rl")
		tui.stdoutViewer.tview.SetText(HelpDocumentation)
		tui.commandInput.tview.SetLabelColor(tcell.ColorGreen)
		tui.commandInput.tview.SetLabel(PROMPT_HELP)
	} else if mode == CommandMode {
		tui.helpBar.tview.SetText(HELP_COMMAND)
		tui.commandInput.tview.SetLabel(PROMPT_CMD)
	}
}

// Create the command-preview element; this will show what the user is actually executing
func NewCommandPreview(execute *string) *TUICommandPreview {
	part := tview.NewTextView().
		SetTextColor(tcell.ColorDefault).
		SetText("rl: " + "[::r]" + *execute + "[-:-:-]").
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

func NewLatencyViewer() *TUILatencyViewer {
	part := tview.NewTextView().
		SetText("").
		SetTextColor(tcell.ColorDefault).
		SetDynamicColors(true)

	return &TUILatencyViewer{part}
}

// Create RL tview application
func NewRLApp(tui *TUI) *TUIApp {
	// -- declare Tview application --
	app := tview.NewApplication()
	// -- we actually want to do the opposite; this prevents mousewheel scroll from breaking things.
	app.EnableMouse(true)
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		return nil, 0
	})

	// if key-event filtering is needed, it can be applied here
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Ctrl + P is intercepted by VSCode; Ctrl + O is the next best thing
		if event.Key() == tcell.KeyCtrlO {
			// TODO
		}
		if event.Key() == tcell.KeyCtrlC {
			tui.Stop()
			tui.chans.exitCode <- 0
			return nil
		}
		return event
	})

	return &TUIApp{app}
}

func NewTextViewer(tui *TUI) *TUITextViewer {
	part := tview.NewTextView().
		SetText(DefaultViewerText).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorDefault)

	onInput := func(event *tcell.EventKey) *tcell.EventKey {
		if tui.stdoutViewer.withDefault {
			// TODO
		}

		switch event.Rune() {
		case ':':
			tui.SetMode(CommandMode)
			return nil
		case '?':
			tui.SetMode(HelpMode)
			return nil
		case '/':
			tui.SetMode(EditMode)
			return nil
		case 'g', 'G':
			// TODO broken and dumb.

			tui.UpdateScrollPosition()
			return event
		}

		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn:
			tui.UpdateScrollPosition()
			return event
		}

		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			if tui.mode == HelpMode {
				// quit back to view-mode
				tui.SetMode(ViewMode)
				// update the line-count
				tui.UpdateScrollPosition()
				return nil
			} else {
				tui.Stop()
				tui.chans.exitCode <- 0
				return nil
			}
		}

		return event
	}

	part.SetInputCapture(onInput)

	return &TUITextViewer{
		tview:       part,
		withDefault: true,
	}
}

func NewCommandInput(tui *TUI) *TUICommandInput {
	ctx := tui.ctx
	cfg := tui.cfg
	state := *tui.state
	execute := ctx.execute

	// When the input is "done" according to tview is when KeyEnter, KeyEscape,
	// KeyTab, KeyDown, KeyUp, KeyBacktab are entered. We can wire custom behaviours into these
	// rather than just terminate the entire app! Normally, done means switch to view mode.
	onDone := func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			tui.state.lineBuffer.SetDone()
			state, _ = state.HandleUserUpdate(tui)
		case tcell.KeyUp:
			tui.ScrollHistoryBack()
			//tui.SetMode(ViewMode)
			//tui.UpdateScrollPosition()
		case tcell.KeyDown:
			tui.ScrollHistoryForward()

			//tui.SetMode(ViewMode)
			//tui.UpdateScrollPosition()
		case tcell.KeyEscape:
			tui.SetMode(ViewMode)
		}
	}

	run := false

	// TODO implement ctrl+left, ctrl+right
	onChange := func(text string) {
		if !run {
			tui.stdoutViewer.tview.SetTextAlign(tview.AlignLeft)
			run = true
		}

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

		tui.commandPreview.UpdateText(*execute, state.lineBuffer, &ctx.envVars)
	}

	commandInput := tview.NewInputField()

	commandInput.
		SetLabelColor(tcell.ColorRed).
		SetChangedFunc(onChange).
		SetLabel(PROMPT_EDIT).
		SetDoneFunc(onDone).
		Focus(func(self tview.Primitive) {
			tui.InvertCommandInput()
		})

	return &TUICommandInput{commandInput}
}

func NewHelpBar(tui *TUI) *TUIHelpBar {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetText(HELP_EDIT)

	return &TUIHelpBar{view}
}

func NewUI(state LineChangeState, cfg *ConfigOpts, ctx *LineChangeCtx, histChan chan *History) *TUI {
	execute := ctx.execute

	tui := TUI{}
	tui.mode = EditMode
	tui.state = &state
	tui.cfg = cfg
	tui.ctx = ctx
	tui.textAlign = tview.AlignCenter

	tui.SetTheme()
	tui.chans.history = histChan
	tui.chans.exitCode = make(chan int, 100)

	tui.app = NewRLApp(&tui)
	tui.latency = NewLatencyViewer()
	tui.commandPreview = NewCommandPreview(execute)
	tui.linePosition = NewLinePosition()
	tui.stdoutViewer = NewTextViewer(&tui)
	tui.commandInput = NewCommandInput(&tui)
	tui.helpBar = NewHelpBar(&tui)

	tui.InvertCommandInput()

	return &tui
}
