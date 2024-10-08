package main

import (
	lol "clx/mods/lol"
	"fmt"
)

type mode string

func (m mode) Run(args []string, module string) {
	fmt.Println("module: ", module)
	if module != "" {
		fmt.Println("qweqwe")
	}
	fmt.Println(lol.Testfunc(args[0]))
}

// экспортируется как символ с именем "Mode"
var Mode mode
