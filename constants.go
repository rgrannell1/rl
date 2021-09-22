package main

const ENVAR_NAME_RL_INPUT = "RL_INPUT" // The environmental-variable name provided to the subcommand passed to execute
const STDIN_BUFFER_SIZE = 100_000_000  // The size of the stdin buffer, in bytes
const USER_WRITE_OCTAL = 00200         // User write file permissions for a file
const USER_READ_WRITE_OCTAL = 0600     // User read-write file permissions for a file

type PromptMode int

const (
	CommandMode PromptMode = iota
	ViewMode
	HelpMode
)

const PROMPT_CMD = "Command | > " // The RL prompt for viewing text
const PROMPT_VIEW = "View    | "  // The RL prompt for executing a command
const PROMPT_HELP = "Help    | "  // The RL prompt for showing help

const HELP_CMD = "Press [green]ESCAPE[-:-:-] to switch to view mode, [green]ENTER[-:-:-] to exit with command-output"
const HELP_VIEW = "Press [green]ESCAPE[-:-:-] or  [green]q[-:-:-] to quit, [green]/[-:-:-] to switch to edit input, [green]?[-:-:-] for help"
const HELP_HELP = "Press [green]ESCAPE[-:-:-] or  [green]q[-:-:-] to quit, [green]/[-:-:-] to switch to edit input"

const DefaultViewerText = `
RL - run commands on key-stroke

Type to replace "$RL_INPUT" with whatever you've typed.
Press [green]ESCAPE[reset] to open view-mode, twice to exit
Press [green]ENTER[reset] to exit with output

Please read RL's "Please Be Careful" documentation
`

const ModesDocumentation = `
RL supports several "modes": command-mode, view-mode, and help-mode.

Command-Mode:

  Run commands on key-stroke. Takes the command you provided as an argument, and substitutes
  $RL_INPUT with whatever you type in. Output
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
  $SHELL           rl starts a command in the user's default-shell.
  $RL_INPUT        this variable conwtains the user-input text. Subcommands
  must use this environmental variable to access user-input.
  <env_vars...>    additional variables provided to rl
`

const Configuration = `
Configuration:
  ~/.config/rl.yaml    RL can be configured in this YAML file. The options are:

  save_history    a boolean value. Should command-execution history be saved to a history file?
                    Defaults to false.

`

const Options = `
Arguments:
  <env_vars>...                          a list of environmental-variable bindings to provide in addition to $RL_INPUT,
                                           of the form "FOO=BAR". This can make it easier to wrap "rl" in an alias or function
                                           wrapper. For example, you could provide "folder=$1" from a bash-function and the
                                           variable "$folder" would be available to the supplied command to search or list.
  <cmd>                                  execute a utility command whenever user input changes; the current line will
                                         be available as the line $RL_INPUT

Options:
  -i, --input-only                       by default,
                                            rl will return its last utility-command execution to standard-output.
                                            When --input-only is enabled, the entered text is returned instead of
                                            the last command's output. This is useful when the utility being
                                            executed is a preview command
  --danger-zone                          run commands without RL attempting to find dangerous user decisions that might
                                           cause unintented system-destruction. Rl can only spot some dangerous usage; the
                                           responsibility to use rl carefully lies with you, with or without
                                           danger-zone enabled. See "Please Be Careful" section of the documentation for
                                           more information.
  - h, --help                            show this documentation
`

const PleaseBeCareful = `
Please Be Careful:
  It is easy to accidentally destroy a system using shell normally; for example, 'rm -rf $FOLDER_NAME' will wipe everything
  in your working-directory if $FOLDER_NAME is empty. Normally, you have the safeguard of at least pressing enter before a
  command is run, giving you time to spot dangerous code. Rl runs its command _every keystroke_, so please think carefully
  about how you use it and which command (and syntax) you provide to <cmd>.

  My recommendations are:
  - Use rl for listing, filtering, selecting, and searching, but never for deleting, updating, moving, or altering
  - Use user-input ($RL_INPUT) exclusively for string arguments to other commands
  - Do not "evaluate" RL_INPUT, especially not in shell. for example, do not invoke bash with a command-argument $RL_INPUT or call eval
  - Always quote $RL_INPUT to avoid word-expansion; this could lead to unexpected evaluation
  - Do not assume you are vigilant enough to ignore these warnings; people fuck up, often.

  RL includes some safety-nets to avoid you running into these problems blindly, but it's not omniscience. Use rl for grep, awk, sed,
  jq, fdfind, and other filtering operations and it will speed up your workflow; use it for rm and it'll uninstall itself (and everything
  else on your system) eventually.
`

const DescriptionDocs = `
rl (readline) is an interactive line-editor.
`

const UsageLine = `
rl
Usage:
  rl [-i|--input-only] [--danger-zone] <cmd> [<env_vars>...]
  rl (-r|--rerun) [--danger-zone]
  rl (-h|--help)
`

const Usage = UsageLine +
	DescriptionDocs +
	ModesDocumentation +
	OutputDocumentation +
	Options +
	PleaseBeCareful +
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
	EnvironmentalVariables +
	PleaseBeCareful

const COMMAND_AND_LINE_ROWS = 2
const STDOUT_ROWS = 0
const SPACE_ROWS = 1
const HELP_ROWS = 1
const COMMAND_ROWS = 1

const ROW_0 = 0
const ROW_1 = 1
const ROW_2 = 2
const ROW_3 = 3
const ROW_4 = 4

const COL_0 = 0
const COL_1 = 1

const ROWSPAN_1 = 1
const COLSPAN_1 = 1
const COLSPAN_2 = 2

const MINHEIGHT_0 = 0

const MINWIDTH_0 = 0
const MINWIDTH_1 = 1

const FOCUS = true
const DONT_FOCUS = false
