package main

import "github.com/docopt/docopt-go"

// Start the interactive line-editor with any provided CLI arguments; execute
// the RL app as a whole
func RL(opts docopt.Opts) int {
	cfg, code := ValidateConfig()
	if code != 0 {
		return code
	}

	state, ctx, code := RLState(&opts)
	if code != 0 {
		return code
	}

	histChan := StartHistoryWriter(cfg)
	defer func() {
		close(histChan)
	}()

	return ctx.CreateUI(state, *cfg, histChan)
}
