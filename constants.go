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

const HelpDocumentation = `
RL

Rl is an interactive command-runner

For questions, feature, bug, or documentation tickets use

https://github.com/rgrannell1/rl/issues

` + ModesDocumentation
