package main

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

	return 0
}
