package main

import (
	lol "clx/mods/lol"
	"fmt"
)

type mode string

func (m mode) Run(args []string, module string) {
	if module != "" {
		fmt.Println("module: ", module)
	}
	fmt.Println(lol.Testfunc(args[0]))
}

// экспортируется как символ с именем "Mode"
var Mode mode
