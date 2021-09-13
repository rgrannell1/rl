package main

import (
	"os"
)

// Start the interactive line-editor with any provided CLI arguments; execute
// the RL app as a whole
func RL(inputOnly bool, rerun bool, execute *string) int {
	shell, code := ReadShell()
	if code != 0 {
		return code
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
