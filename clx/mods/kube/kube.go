package main

import (
	modules "clx/mods/kube/modules"
	utils "clx/utils"
	"crypto/tls"
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

func checkKube(target string, wg *sync.WaitGroup, sem chan struct{}) {
	defer func() {
		<-sem
		wg.Done()
	}()
	ports := map[string][]string{"kubeapi": {"6443"}, "kubelet": {"10250"}}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.Client{
		Timeout: 1 * time.Second,
	}

	//check kubeapi
	for _, port := range ports["kubeapi"] {
		url := fmt.Sprintf("https://%s:%s", target, port)
		req, _ := http.NewRequest(http.MethodGet, url, nil)

		resp, err := client.Do(req)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
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
		req, _ := http.NewRequest(http.MethodGet, url, nil)

		resp, err := client.Do(req)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
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

	moduleName, err := utils.GetParam(args, "-M")
	if err != nil {
		fmt.Println("You have to chose module here")
		os.Exit(0)
	}

	//Mode logic
	if moduleName == "" {
		var wg sync.WaitGroup
		sem := make(chan struct{}, 256)
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
