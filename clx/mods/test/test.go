package main

import (
	lol "clx/mods/test/lol"
	"fmt"
)

type mode string

func (m mode) Run(args []string) {
	module := ""
	for i, arg := range args {
		if arg == "-M" {
			module = args[i+1]
		}
	}

	if module != "" {
		fmt.Println("module: ", module)
	}
	fmt.Println(lol.Testfunc(args[0]))
}

// экспортируется как символ с именем "Mode"
var Mode mode
