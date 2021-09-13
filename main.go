package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docopt/docopt-go"
)

func main() {
	content, err := ioutil.ReadFile("cli-document.txt")

	if err != nil {
		fmt.Printf("RL: failed to read usage-information from cli-document.txt; cannot launch CLI")
		os.Exit(1)
	}

	usage := string(content)
	opts, _ := docopt.ParseDoc(usage)

	execute, execErr := opts.String("--execute")

	if execErr != nil {
		execute = ""
	}

	input, inputErr := opts.Bool("--input-only")

	if inputErr != nil {
		fmt.Printf("RL: failed to read --input-only option. %v\n", inputErr)
		os.Exit(1)
	}

	rerun, rerunErr := opts.Bool("--rerun")

	if rerunErr != nil {
		fmt.Printf("RL: failed to read --rerun option. %v\n", rerunErr)
		os.Exit(1)
	}

	os.Exit(RL(input, rerun, &execute))
}
