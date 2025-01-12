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

func getFlags(args []string) map[string]string {
	requiredParams := map[string]string{
		"-M": "module",
		"-t": "threads",
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
			go checkKube(target, &wg, sem)

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
