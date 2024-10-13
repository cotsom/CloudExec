package main

import (
	modules "clx/mods/test/modules"
	utils "clx/mods/test/utils"
	"fmt"
	"os"
	"path/filepath"
)

type mode string

func (m mode) Run(args []string) {
	if len(args) < 1 {
		fmt.Println("Enter host or subnetwork")
		return
	}

	module := utils.GetParamModule(args)

	if module != "" {
		rootPath, err := os.Executable()
		if err != nil {
			fmt.Println(err)
			return
		}
		modules, err := utils.GetModulesName(fmt.Sprintf("%s/mods/test/modules", filepath.Dir(rootPath)))
		if err != nil {
			fmt.Println(err)
			return
		}

		if !utils.Contains(modules, module) {
			fmt.Printf("there is no %s module, chose ones from list \n%s\n", module, modules)
			return
		}

		// fmt.Println("module: ", module)
	}
	fmt.Println(modules.Testfunc(args[0]))
}

// экспортируется как символ с именем "Mode"
var Mode mode
