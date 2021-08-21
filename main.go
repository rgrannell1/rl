package main

import (
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
)

func main() {
	usage := `rl
Usage:
	rl [-s|--show-all]
	rl [-s|--show-all] [-x <cmd>|--execute <cmd>] [-i|--input-only]
	rl (-h|--help)

Description:
  rl (readline) is an interactive line-editor

Options:
	-s, --show-all                         by default rl clears the terminal after each keypress and before utility execution; provide -s to suppress this and keep all output present
	-i, --input-only                       redundant if not running in --execute mode. by default, rl will return its last utility-command execution to standard-output. When --input-only is enabled, the entered text is returned instead of the last command's output. This is useful when the utility being executed is a preview command
	-x <command>, --execute <command>      execute a utility command whenever user input changes; the current line will be available as the line $RL_INPUT
	- h, --help                            show this documentation

Environment Variables:
	$SHELL       when run with -x or --execute, rl starts a command in the user's default-shell.
	$RL_INPUT    when run with -x or --execute, this variable contains the user-input text. Subcommands must use this environmental variable to access user-input.

See Also:
  fzf, selecta, dmenu, percol

License:
	The MIT License

	Copyright (c) 2021 Róisín Grannell

	Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`
	opts, _ := docopt.ParseDoc(usage)
	show, showErr := opts.Bool("--show-all")

	if showErr != nil {
		fmt.Printf("RL: failed to read show option. %v\n", showErr)
		os.Exit(1)
	}

	execute, execErr := opts.String("--execute")

	if execErr != nil {
		execute = ""
	}

	input, inputErr := opts.Bool("--input-only")

	if inputErr != nil {
		fmt.Printf("RL: failed to read --input-only option. %v\n", showErr)
		os.Exit(1)
	} else if input && execErr != nil {
		fmt.Printf("RL: do not provide --input-only option without specifying a command using -x or --execute. %v\n", showErr)
		os.Exit(1)
	}

	rl(show, input, &execute)
}
