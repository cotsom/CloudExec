package main

import (
	modules "clx/mods/test/modules"
	utils "clx/mods/test/utils"
	"fmt"
)

type mode string

func (m mode) Run(args []string) {
	module := ""
	for i, arg := range args {
		if arg == "-M" {
			if len(args) != i+1 {
				module = args[i+1]
			} else {
				fmt.Println("You have to chose module here")
			}
		}
	}

	if module != "" {
		modules, err := utils.GetModulesName("modules")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("modules: ", modules)
		if !utils.Contains(modules, module) {
			fmt.Sprintf("there is no %s module, chose ones from list %s", module, modules)
		}

		// fmt.Println("module: ", module)
	}
	fmt.Println(modules.Testfunc(args[0]))
}

// экспортируется как символ с именем "Mode"
var Mode mode
