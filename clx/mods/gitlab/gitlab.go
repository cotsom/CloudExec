package main

import (
	modules "clx/mods/gitlab/modules"
	utils "clx/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// mode type for plugin's symbol
type mode string

type Module interface {
	RunModule(target string, flags map[string]string, scheme string)
}

var registeredModules = map[string]Module{
	"loginbypass": modules.Loginbypass{},
	// Add another modules here
}

func getFlags(args []string) map[string]string {
	requiredParams := map[string]string{
		"-M":     "module",
		"-t":     "threads",
		"--port": "port",
		"-u":     "user",
		"-p":     "password",
		"-iL":    "inputlist",
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

func checkGitlab(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	scheme := "http"
	gitlabRoute := "users/sign_in"

	if flags["port"] == "" {
		flags["port"] = "80"
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	// Make http req
	url := fmt.Sprintf("http://%s:%s/%s", target, flags["port"], gitlabRoute)

	response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		return
	}
	defer response.Body.Close()
	respBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
	}

	// Make https req
	if strings.Contains(string(respBody), "HTTP request was sent to HTTPS port") {
		url = fmt.Sprintf("https://%s:%s/%s", target, flags["port"], gitlabRoute)
		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			return
		}
		defer response.Body.Close()
		respBody, err = ioutil.ReadAll(response.Body)

		if err != nil {
			// fmt.Printf("client: could not read response body: %s\n", err)
			return
		}
		scheme = "https"
	}

	// fmt.Println(string(respBody))
	if !strings.Contains(string(respBody), "gitlab") {
		return
	}
	utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Gitlab\n", utils.ClearLine, target, flags["port"]))

	if flags["module"] != "" {
		if module, exists := registeredModules[flags["module"]]; exists {
			module.RunModule(target, flags, scheme)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", flags["module"], registeredModules)
			os.Exit(1)
		}
	}

}

// Main func
func (m mode) Run(args []string) {
	if len(args) < 1 {
		fmt.Println("Enter host or subnetwork")
		return
	}

	flags := getFlags(args)
	var targets []string
	if flags["inputlist"] != "" {
		targets = utils.ParseTargetsFromList(flags["inputlist"])
	} else {
		targets = utils.ParseTargets(args[0])
	}

	//MAIN LOGIC
	var wg sync.WaitGroup
	var sem chan struct{}

	//set threads
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

	progress := 0
	for i, target := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go checkGitlab(target, &wg, sem, flags)
		utils.ProgressBar(len(targets), i+1, &progress)
	}
	fmt.Println("")
	wg.Wait()

}

// exporting symbol "Mode"
var Mode mode
