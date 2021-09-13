package main

import (
	"fmt"
	"os"
)

// Start the interactive line-editor with any provided CLI arguments
func RL(inputOnly bool, execute *string) int {
	shell := os.Getenv("SHELL")
	if shell == "" {
		fmt.Printf("RL: could not determine user's shell (e.g bash, zsh). Ensure $SHELL is set.")
		return 1
	}

	stdin, code := ReadStdin()
	if code != 0 {
		return code
	}

	cfg, code := ValidateConfig()
	if code != 0 {
		return code
	}

	ctx := LineChangeCtx{
		shell,
		inputOnly,
		execute,
		os.Environ(),
		nil,
		stdin,
	}

	linebuffer := LineBuffer{}
	state := LineChangeState{
		lineBuffer: &linebuffer,
		cmd:        nil,
	}

	histChan := StartHistoryWriter(cfg)
	defer func() {
		close(histChan)
	}()

	return ctx.CreateUI(state, *cfg, histChan, execute)
}
