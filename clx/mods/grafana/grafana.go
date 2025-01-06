package main

import (
	modules "clx/mods/test/modules"
	utils "clx/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

func getFlags(args []string) map[string]string {
	requiredParams := map[string]string{
		"-M": "module",
		"-t": "threads",
		"-p": "ports",
	}

	flags := make(map[string]string)

	for key, name := range requiredParams {
		value, err := utils.GetParam(args, key)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		flags[name] = value
	}

	return flags
}

func checkGrafana(target string, wg *sync.WaitGroup, sem chan struct{}, port string) {
	// fmt.Println(target)
	defer func() {
		<-sem
		wg.Done()
	}()

	if port == "" {
		port = "3000"
	}

	grafanaDefaultCreds := [3]string{"admin:admin", "admin:prom-operator", "admin:openbmp"}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	//check grafana port
	url := fmt.Sprintf("http://%s:%s", target, port)

	response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		return
	}
	respBody, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
	}

	if strings.Contains(string(respBody), "grafana") {
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - Grafana found!", target))
		for _, creds := range grafanaDefaultCreds {
			url := fmt.Sprintf("http://%s@%s:%s/api/datasources", creds, target, port)
			response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
			if err != nil {
				fmt.Println(err)
			}

			if response.StatusCode == 200 {
				utils.Colorize(utils.ColorGreen, fmt.Sprintf("[*] %s - Default creds found! (%s)", target, creds))
			}
		}
	}
}

// Main func
func (m mode) Run(args []string) {
	if len(args) < 1 {
		fmt.Println("Enter host or subnetwork")
		return
	}

	var targets = utils.ParseTargets(args[0])

	flags := getFlags(args)

	//Mode logic
	if flags["module"] == "" {
		var wg sync.WaitGroup
		var sem chan struct{}

		if flags["threads"] != "" {
			threads, err := strconv.Atoi(flags["threads"])
			if err != nil {
				fmt.Println("You have to set correct number of threads")
				os.Exit(0)
			}
			sem = make(chan struct{}, threads)
		} else {
			sem = make(chan struct{}, 100)
		}
		for _, target := range targets {
			wg.Add(1)
			sem <- struct{}{}
			go checkGrafana(target.String(), &wg, sem, flags["ports"])
		}
		wg.Wait()
	} else {
		rootPath, err := os.Executable()
		if err != nil {
			fmt.Println(err)
			return
		}
		modules := utils.GetModulesName(fmt.Sprintf("%s/mods/test/modules", filepath.Dir(rootPath)))

		if !utils.Contains(modules, flags["module"]) {
			fmt.Printf("there is no %s module, chose ones from list \n%s\n", flags["module"], modules)
			return
		}

		if module, exists := registeredModules[flags["module"]]; exists {
			module.RunModule("lol")
		} else {
			fmt.Printf("Module %s not found. Available modules: %v\n", flags["module"], modules)
			os.Exit(1)
		}
	}

}

// exporting symbol "Mode"
var Mode mode
