package main

import (
	modules "clx/mods/kube/modules"
	utils "clx/utils"
	"fmt"
	"os"
	"path/filepath"
)

// mode type for plugin's symbol
type mode string

type Module interface {
	RunModule(arg string)
}

var registeredModules = map[string]Module{
	"module1": modules.Module1{},
	// Add another modules here
}

func (m mode) Run(args []string) {
	if len(args) < 1 {
		fmt.Println("Enter host or subnetwork")
		return
	}

	moduleName := utils.GetParamModule(args)

	if moduleName != "" {
		rootPath, err := os.Executable()
		if err != nil {
			fmt.Println(err)
			return
		}

		modules := utils.GetModulesName(fmt.Sprintf("%s/mods/kube/modules", filepath.Dir(rootPath)))

		if !utils.Contains(modules, moduleName) {
			fmt.Printf("there is no %s module, chose ones from list \n%s\n", moduleName, modules)
			return
		}

		if module, exists := registeredModules[moduleName]; exists {
			module.RunModule("lol")
		} else {
			fmt.Printf("Module %s not found. Available modules: %v\n", moduleName, modules)
			os.Exit(1)
		}
	}

	fmt.Println("working with", args[0], "here")
}

// exporting symbol "Mode"
var Mode mode
