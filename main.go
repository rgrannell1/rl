package main

import (
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
)

func main() {
	usage := `rl
Usage:
	rl [-c|--clear] [-x <cmd>|--execute <cmd>]

Description:
  rl (readline) is an interactive line editor.

	It captures keypresses and backspaces (to remove characters), and terminates when
	escape, enter, or ctrl + c is pressed. By default, it echos its line buffer for each keypress, as shown below:

	> rl

	h
	he
	hel
	hell
	hello

	setting --clear will prompt rl to clear the terminal (/dev/tty) after each keypress.

Options:
	-c, --clear                          clear the terminal after each update.
	-x <command>, --execute <command>    execute a command on readline change; the current line will be available as the line $RL_INPUT

License:
	The MIT License

	Copyright (c) 2021 Róisín Grannell

	Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`
	opts, _ := docopt.ParseDoc(usage)
	clear, clearErr := opts.Bool("--clear")

	if clearErr != nil {
		fmt.Printf("RL: failed to read clear option. %v\n", clearErr)
		os.Exit(1)
	}

	execute, execErr := opts.String("--execute")

	if execErr != nil {
		execute = ""
	}

	rl(clear, &execute)
}
