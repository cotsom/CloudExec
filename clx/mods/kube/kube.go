package main

import (
	modules "clx/mods/kube/modules"
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

func checkKube(target string, wg *sync.WaitGroup, sem chan struct{}) {
	fmt.Println(target)
	defer func() {
		<-sem
		wg.Done()
	}()
	ports := map[string][]string{"kubeapi": {"6443"}, "kubelet": {"10250"}}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	//check kubeapi
	for _, port := range ports["kubeapi"] {
		url := fmt.Sprintf("https://%s:%s", target, port)

		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		defer response.Body.Close()

		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		if strings.Contains(string(respBody), "\"apiVersion\"") {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - kube Api", target))
		}
	}

	//check kubelet
	for _, port := range ports["kubelet"] {
		url := fmt.Sprintf("https://%s:%s/pods", target, port)
		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		defer response.Body.Close()

		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		if strings.Contains(string(respBody), "Unauthorized") {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - kubelet", target))
		} else {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - kubelet UNAUTH!", target))
		}
	}
}

func (m mode) Run(args []string) {
	//Get input data
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
			go checkKube(target.String(), &wg, sem)

		}
		wg.Wait()
		// close(sem)
	} else {
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

}

// exporting symbol "Mode"
var Mode mode
