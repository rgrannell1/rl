package main

import (
	"os"

	"github.com/docopt/docopt-go"
)

func main() {
	opts, err := docopt.ParseDoc(Usage)

	if err != nil {
		panic(err)
	}

	os.Exit(RL(opts))
}
