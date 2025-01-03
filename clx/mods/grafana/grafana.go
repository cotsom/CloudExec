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

func checkGrafana(target string, wg *sync.WaitGroup, sem chan struct{}) {
	defer func() {
		<-sem
		wg.Done()
	}()
	port := "3000"
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
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - Grafana", target))
		for _, creds := range grafanaDefaultCreds {
			url := fmt.Sprintf("http://%s@%s:%s", creds, target, port)
			response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
			if err != nil {
				fmt.Println(err)
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

	//Get module
	moduleName, err := utils.GetParam(args, "-M")
	if err != nil {
		fmt.Println("You have to chose module here")
		os.Exit(0)
	}

	//Get threads number
	threadsNumber, err := utils.GetParam(args, "-t")
	if err != nil {
		fmt.Println("You have to set threads number")
		os.Exit(0)
	}

	//Mode logic
	if moduleName == "" {
		var wg sync.WaitGroup
		var sem chan struct{}

		if threadsNumber != "" {
			threads, err := strconv.Atoi(threadsNumber)
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
			go checkGrafana(target.String(), &wg, sem)
			// time.Sleep(time.Second * 2)
		}
		wg.Wait()
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
