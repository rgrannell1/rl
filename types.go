package main

import (
	"os/exec"
	"time"

	"github.com/rivo/tview"
)

//  Stores user-input text, and whether a terminal character has been reached.
type LineBuffer struct {
	// The user-entered characters
	content string
	// Has a terminal character been reached?
	done bool
}

// Store variables that will changes as characters are received from the user
// and commands are executed.
type LineChangeState struct {
	lineBuffer *LineBuffer // a pointer to an array of characters the user has entered into this application, excluding some special characters like backspaces.
	cmd        *exec.Cmd   // a pointer to the command currently being executed, if rl is running in execute mode
}

// Contextual contantish information like the user's shell, environmental variables, and command-line options
type LineChangeCtx struct {
	shell       string          // the user's shell-variable
	inputOnly   bool            // should we only return the user's input (e.g lineBuffer) instead of the final command execution, if we're running in execute mode?
	execute     *string         // a string to execute in a user's shell
	environment []string        // an array of this processes environmental variables
	tgt         *tview.TextView // where to pipe output
}

// RL Configuration structure
type ConfigOpts struct {
	HistoryPath string // the history path for RL
	ConfigPath  string // the config path for RL
	Config      RLConfigFile
}

type RLConfigFile struct {
	SaveHistory bool `yaml:"save_history"`
}

type History struct {
	Input     string    `json:"input"`
	Command   string    `json:"command"`
	Template  string    `json:"template"`
	Time      time.Time `json:"time"`
	StartTime time.Time `json:"start_time"`
}
