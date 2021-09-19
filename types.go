package main

import (
	"bufio"
	"os"
	"os/exec"
	"time"

	"github.com/smallnest/ringbuffer"
)

//  Stores user-input text, and whether a terminal character has been reached.
type LineBuffer struct {
	content string // The user-entered character?
	done    bool   // Has a terminal character been reached?
}

func (buff *LineBuffer) SetDone() *LineBuffer {
	buff.done = true
	return buff
}

// Store variables that will changes as characters are received from the user
// and commands are executed.
type LineChangeState struct {
	lineBuffer *LineBuffer // a pointer to an array of characters the user has entered into this application, excluding some special characters like backspaces.
	cmd        *exec.Cmd   // a pointer to the command currently being executed, if rl is running in execute mode
}

// Contextual contantish information like the user's shell, environmental variables, and command-line options
type LineChangeCtx struct {
	shell       string                 // the user's shell-variable
	inputOnly   bool                   // should we only return the user's input (e.g lineBuffer) instead of the final command execution, if we're running in execute mode?
	execute     *string                // a string to execute in a user's shell
	environment []string               // an array of this processes environmental variables
	stdin       *ringbuffer.RingBuffer // a buffer containing as much stdin as we are willing to store
}

// RL Configuration structure
type ConfigOpts struct {
	HistoryPath string       // the history path for RL
	ConfigPath  string       // the config path for RL
	Config      RLConfigFile // RL configuration
}

// RL Configuration file-data
type RLConfigFile struct {
	SaveHistory bool `yaml:"save_history"` // A configuration option. Should a history-file be used?
}

// RL History Information
type History struct {
	Input     string    `json:"input"`      // The user-entered input text
	Command   string    `json:"command"`    // The command executed
	Template  string    `json:"template"`   // The 'template' the user provided to -x
	Time      time.Time `json:"time"`       // The time the command was started, approximately
	StartTime time.Time `json:"start_time"` // The start-time of the program, approximately. Can be used as an ID.
}

type HistoryCursor struct {
	historyPath string
	index       int
	buffer      ringbuffer.RingBuffer
}

func (curs *HistoryCursor) GetCount() int {
	conn, _ := os.Open(curs.historyPath)

	// Create new Scanner.
	scanner := bufio.NewScanner(conn)

	count := 0
	for scanner.Scan() {
		scanner.Text()
		count += 1
	}

	return count
}

func (curs *HistoryCursor) GetHistory(index int) History {
	return History{}
}
