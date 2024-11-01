package main

import (
	modules "clx/mods/test/modules"
	utils "clx/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

func checkGrafana(target string, wg *sync.WaitGroup, sem chan struct{}) {
	defer func() {
		<-sem
		wg.Done()
	}()
	ports := map[string][]string{"grafana": {"3000"}, "prometheus": {"9090"}}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	//check grafana port
	for _, port := range ports["grafana"] {
		// fmt.Println(target)
		url := fmt.Sprintf("http://%s:%s", target, port)

		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			// fmt.Println(err)
			continue
		}
		respBody, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		if strings.Contains(string(respBody), "grafana") {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - Grafana", target))
			url := fmt.Sprintf("http://admin:admin@%s:%s", target, port)
			response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
			if err != nil {
				// fmt.Println(err)
				continue
			}
			respBody, err := ioutil.ReadAll(response.Body)
			defer response.Body.Close()
			if strings.Contains(string(respBody), "grafana") {

			}
		}
	}
}

func (m mode) Run(args []string) {
	if len(args) < 1 {
		fmt.Println("Enter host or subnetwork")
		return
	}

	var targets = utils.ParseTargets(args[0])
	// fmt.Println(targets)

	moduleName, err := utils.GetParam(args, "-M")
	if err != nil {
		fmt.Println("You have to chose module here")
		os.Exit(0)
	}

	//Mode logic
	if moduleName == "" {
		var wg sync.WaitGroup
		sem := make(chan struct{}, 100)
		for _, target := range targets {
			wg.Add(1)
			sem <- struct{}{}
			go checkGrafana(target.String(), &wg, sem)
			// time.Sleep(time.Second * 2)
		}
	} else {
		rootPath, err := os.Executable()
		if err != nil {
			fmt.Println(err)
			return
		}
		modules := utils.GetModulesName(fmt.Sprintf("%s/mods/test/modules", filepath.Dir(rootPath)))

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

}

// exporting symbol "Mode"
var Mode mode
