package main

import (
	"fmt"
	"strings"
)

/*
Shell parse-tree

	command = [a-zA-Z-_0-9]+

	subcommand = [a-zA-Z-_0-9]+
	arguments = ["'][a-zA-Z-_0-9$]+["']
	flag = [[-]{1,2}a-z+]
	call = command [subcommand] [<argument...>] [<flag...>] [<argument...>]
	combiners = <call> [&&|;|`||`] <call>
	input = argument | call | combiners

Todo
*/
func AuditCommand(command *string) int {
	if strings.Trim(*command, "") == "$RL_INPUT" {
		fmt.Printf("RL: do not use $RL_INPUT in unquoted format; it's dangerous.\n")

		return 1
	}

	return 0
}
