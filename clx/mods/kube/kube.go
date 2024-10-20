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

func checkKubeApi(target string, c chan string) {
	// defer wg.Done()
	ports := [3]string{"6443"}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for _, port := range ports {
		fmt.Println("target:", target, "port:", port)
		client := http.Client{
			Timeout: 3 * time.Second,
		}
		url := fmt.Sprintf("https://%s:%s", target, port)
		req, _ := http.NewRequest(http.MethodGet, url, nil)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		if strings.Contains(string(respBody), "\"apiVersion\"") {
			c <- fmt.Sprintf("[*] %s - kube Api", target)
			// utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - kube Api", target))
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
		// var wg sync.WaitGroup
		c := make(chan string)
		for _, target := range targets {
			// wg.Add(1)
			go checkKubeApi(target.String(), c)
		}
		// go func() {
		// 	wg.Wait()
		// 	close(c)
		// }()

		for result := range c {
			fmt.Println(result)
		}

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
