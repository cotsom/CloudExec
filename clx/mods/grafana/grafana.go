package main

import (
	modules "clx/mods/grafana/modules"
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
	RunModule(target string, flags map[string]string, wg *sync.WaitGroup, sem chan struct{})
}

var registeredModules = map[string]Module{
	"datasource": modules.Datasource{},
	"defcreds":   modules.Defcreds{},
	// Add another modules here
}

func getFlags(args []string) map[string]string {
	requiredParams := map[string]string{
		"-M":     "module",
		"-t":     "threads",
		"--port": "port",
		"-u":     "user",
		"-p":     "password",
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

func checkGrafana(target string, wg *sync.WaitGroup, sem chan struct{}, port string, flags map[string]string, mu *sync.Mutex, foundTargets *[]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	if port == "" {
		port = "3000"
	}

	creds := fmt.Sprintf("%s:%s", flags["user"], flags["password"])

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	//check grafana port
	url := fmt.Sprintf("http://%s:%s", target, port)

	response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		return
	}
	defer response.Body.Close()
	respBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
	}

	if !strings.Contains(string(respBody), "grafana") {
		return
	}

	//Use mutex for write to var
	mu.Lock()
	*foundTargets = append(*foundTargets, target)
	mu.Unlock()

	url = fmt.Sprintf("http://%s@%s:%s/api/datasources", creds, target, port)
	response, err = utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s - Grafana! (%s)\n", utils.ClearLine, target, creds))
	} else {
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s - Grafana\n", utils.ClearLine, target))
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
	var foundTargets []string

	//Main logic
	var wg sync.WaitGroup
	var sem chan struct{}
	var mu sync.Mutex

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
		go checkGrafana(target.String(), &wg, sem, flags["port"], flags, &mu, &foundTargets)
		utils.ProgressBar(len(targets), i+1, &progress)
	}
	wg.Wait()

	//Mode logic
	if flags["module"] != "" {
		if module, exists := registeredModules[flags["module"]]; exists {
			for _, target := range foundTargets {
				wg.Add(1)
				sem <- struct{}{}
				go module.RunModule(target, flags, &wg, sem)
			}
			wg.Wait()
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", flags["module"], registeredModules)
			os.Exit(1)
		}
	}

}

// exporting symbol "Mode"
var Mode mode
