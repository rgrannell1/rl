package main

import (
	"os"
	"os/exec"
)

//  Stores user-input text, and whether a terminal character has been reached.
type LineBuffer struct {
	// The user-entered characters
	runes []rune
	// Has a terminal character been reached?
	done bool
}

// Convert a linebuffer to a string
func (buff *LineBuffer) String() string {
	return string(buff.runes)
}

// Handle a backspace; remove all but the last character in
// the linebuffer runes
func (buff *LineBuffer) Backspace() LineBuffer {
	if len(buff.runes) == 0 {
		return *buff
	} else {
		buff.runes = buff.runes[:len(buff.runes)-1]
		return *buff
	}
}

// Add a character to a linebuffer
func (buff *LineBuffer) AddChar(char rune) LineBuffer {
	buff.runes = append(buff.runes, char)
	return *buff
}

// Store variables that will changes as characters are received from the user
// and commands are executed.
type LineChangeState struct {
	lineBuffer *LineBuffer // a pointer to an array of characters the user has entered into this application, excluding some special characters like backspaces.
	cmd        *exec.Cmd   // a pointer to the command currently being executed, if rl is running in execute mode
}

// Contextual contantish information like the user's shell, environmental variables, and command-line options
type LineChangeCtx struct {
	shell       string   // the user's shell-variable
	tty         *os.File // a pointer to /dev/tty
	show        bool     // is the show option enabled? i.e should we avoid clearing the screen pre-command execution?
	inputOnly   bool     // should we only return the user's input (e.g lineBuffer) instead of the final command execution, if we're running in execute mode?
	execute     *string  // a string to execute in a user's shell
	environment []string // an array of this processes environmental variables
}
