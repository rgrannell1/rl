package main

const ENVAR_NAME_RL_INPUT = "RL_INPUT"                                            // The environmental-variable name provided to the subcommand passed to execute
const STDIN_BUFFER_SIZE = 100000000                                               // The size of the stdin buffer, in bytes
const USER_WRITE_OCTAL = 00200                                                    // Allow user writes, no other permissions
const PROMPT_CMD = "Command | > "                                                 // The RL prompt for input
const PROMPT_VIEW = "View | (Esc or q to exit, '/' to run command, '?' for help)" // The RL prompt for input
const HELP_VIEW = "Help | (Esc or q to exit, '/' to run command)"

type PromptMode int

const (
	CommandPrompt PromptMode = iota
	ViewPrompt
	HelpPrompt
)
