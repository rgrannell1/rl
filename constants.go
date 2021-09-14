package main

const ENVAR_NAME_RL_INPUT = "RL_INPUT"                                            // The environmental-variable name provided to the subcommand passed to execute
const STDIN_BUFFER_SIZE = 100_000_000                                             // The size of the stdin buffer, in bytes
const PROMPT_CMD = "Command | > "                                                 // The RL prompt for viewing text
const PROMPT_VIEW = "View | (Esc or q to exit, '/' to run command, '?' for help)" // The RL prompt for executing a command
const HELP_VIEW = "Help | (Esc or q to exit, '/' to run command)"                 // The RL prompt for showing help
const USER_WRITE_OCTAL = 00200                                                    // User write file permissions for a file
const USER_READ_WRITE_OCTAL = 0600                                                // User read-write file permissions for a file

type PromptMode int

const (
	CommandMode PromptMode = iota
	ViewMode
	HelpMode
)
