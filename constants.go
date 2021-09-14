package main

const ENVAR_NAME_RL_INPUT = "RL_INPUT"                                               // The environmental-variable name provided to the subcommand passed to execute
const STDIN_BUFFER_SIZE = 100_000_000                                                // The size of the stdin buffer, in bytes
const PROMPT_CMD = "Command | > "                                                    // The RL prompt for viewing text
const PROMPT_VIEW = "View    | (Esc or q to exit, '/' to run command, '?' for help)" // The RL prompt for executing a command
const HELP_VIEW = "Help    | (Esc or q to exit, '/' to run command)"                 // The RL prompt for showing help
const USER_WRITE_OCTAL = 00200                                                       // User write file permissions for a file
const USER_READ_WRITE_OCTAL = 0600                                                   // User read-write file permissions for a file

type PromptMode int

const (
	CommandMode PromptMode = iota
	ViewMode
	HelpMode
)

const ModesDocumentation = `
RL supports several "modes": command-mode, view-mode, and help-mode.

Command-Mode:

  Run commands on key-stroke. Takes the command you provided with
  --execute, -x, and substitutes $RL_INPUT with whatever you type in. Output
  (stdout, stderr) is shown on-screen. Press Enter to run the command, output to
  (stdout, stderr), and exit RL.

  If you run into issues with missing output, see 'Output' section of this
  documentation

  The text you've entered is shown in the bottom row of rl. The command you executed
  on keypress is shown in the top-left of the screen.

  - Escape       switch to view mode
  - Enter        output to stdout + stderr and exit
  - Backspace    delete char before cursor
  - Delete       delete char after cursor

  Cursor Navigation:
  - Left, Right              move cursor left, right
  - Home, Ctrl-A, Alt-A      start-of-line
  - End, Ctrl-E, Alt-E       end-of-line
  - Ctrl-Left, Ctrl-Right    move one word left, right

View-Mode:

  Scroll through command-output text. This is useful when a command produces a
  lot of output, for example grepping a log-file. Line-position is shown in the
  top left corner of rl

  - Escape, q    quit without output
  - /            switch to command-mode
  - ?            switch to help-mode

  Text Navigation:

  - Up, k                 scroll up
  - Down, j               scroll down
  - g                     move to top
  - G                     move to bottom
  - Page Up, Page down    scroll faster
`

const OutputDocumentation = `
Output:

  Commands like grep, awk, sed, and jq buffer their output

`

const License = `
License:
  The MIT License

  Copyright (c) 2021 Róisín Grannell

  Permission is hereby granted, free of charge, to any person obtaining a copy of this
  software and associated documentation files (the "Software"), todeal in the Software
  without restriction, including without limitation the rights to use, copy, modify, merge,
  publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons
  to whom the Software is furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all copies or
  substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
  BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,FITNESS FOR A PARTICULAR PURPOSE AND
  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`

const SeeAlso = `
See Also:
  fzf, selecta, dmenu, percol

`

const HistoryDocs = `
History
  ~/.local/share/rl/history    If enabled, RL will save each executed command to a history file
                               in JSON format.

`

const EnvironmentalVariables = `
Environment Variables:
  $SHELL       when run with -x or --execute, rl starts a command in the user's default-shell.
  $RL_INPUT    when run with -x or --execute, this variable contains the user-input text. Subcommands
  must use this environmental variable to access
                user-input.
`

const Configuration = `
Configuration:
  ~/.config/rl.yaml    RL can be configured in this YAML file. The options are:

  save_history    a boolean value. Should command-execution history be saved to a history file?
                    Defaults to false.

`

const Options = `
Options:
  -i, --input-only                       redundant if not running in --execute mode. by default,
                                            rl will return its last utility-command execution to standard-output.
                                            When --input-only is enabled, the entered text is returned instead of
                                            the last command's output. This is useful when the utility being
                                           executed is a preview command
  -x <command>, --execute <command>      execute a utility command whenever user input changes; the current line will
                                           be available as the line $RL_INPUT
  --danger-zone                          run commands with no validation; allows commands like 'rm' to execute unchecked.
  - h, --help                            show this documentation
`

const DescriptionDocs = `
rl (readline) is an interactive line-editor.
`

const Usage = `
rl
Usage:
  rl [-x <cmd>|--execute <cmd>] [-i|--input-only] [--danger-zone]
  rl (-r|--rerun) [--danger-zone]
  rl (-h|--help)
` +
	DescriptionDocs +
	ModesDocumentation +
	OutputDocumentation +
	Options +
	Configuration +
	HistoryDocs +
	EnvironmentalVariables +
	License +
	SeeAlso

const HelpDocumentation = `
RL

Rl is an interactive command-runner

For questions, feature, bug, or documentation tickets use

https://github.com/rgrannell1/rl/issues

` +
	ModesDocumentation +
	OutputDocumentation +
	Configuration +
	HistoryDocs +
	EnvironmentalVariables
