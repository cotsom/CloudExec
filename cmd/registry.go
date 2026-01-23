/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	modules "github.com/cotsom/CloudExec/internal/modules/registry"
	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Module interface {
	RunModule(target string, flags map[string]string, scheme string)
}

var registeredModules = map[string]Module{
	"images": modules.Images{},
	"harbor": modules.Harbor{},
	// Add another modules here
}

// registryCmd represents the registry command
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "discover & exploit container registry",
	Long: `Mode for discover & exploit Kafka
Will scan and highlight all found hosts with kafka.

Modules:
* images - Get all images from docker registry
* harbor - Get all artifacts from Harbor registry (containers build history, helm charts values.yml)`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := make(map[string]string)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			flags[f.Name] = f.Value.String()
		})

		targets, err := utils.GetTargets(flags, args)
		if err != nil {
			utils.Colorize(utils.ColorRed, err.Error())
			return
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
			go checkRegistry(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)

	registryCmd.Flags().IntP("threads", "t", 100, "threads")
	registryCmd.Flags().StringP("port", "", "", "registry port")
	registryCmd.Flags().StringP("user", "u", "", "registry user")
	registryCmd.Flags().StringP("password", "p", "", "registry password")
	registryCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	registryCmd.Flags().StringP("module", "M", "", "Choose module")
	registryCmd.Flags().StringP("timeout", "", "", "Count of seconds for waiting http response")
}

func checkRegistry(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	port, err := utils.SetPort(flags["port"], "5000")
	if err != nil {
		utils.Colorize(utils.ColorRed, err.Error())
		return
	}

	scheme := "http"
	if port == "443" {
		scheme = "https"
	}

	regitryRoute := "v2/_catalog"
	creds := fmt.Sprintf("%s:%s", flags["user"], url.QueryEscape(flags["password"]))

	if flags["timeout"] == "" {
		flags["timeout"] = "1"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Make http req
	url := fmt.Sprintf("http://%s@%s:%s/%s", creds, target, port, regitryRoute)

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
		url = fmt.Sprintf("https://%s@%s:%s/%s", creds, target, port, regitryRoute)
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
	if response.StatusCode == 200 {
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Registry\n", utils.ClearLine, target, port))
	} else if response.StatusCode != 404 {
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Registry\n", utils.ClearLine, target, port))
	}

	if flags["module"] != "" {
		if module, exists := registeredModules[flags["module"]]; exists {
			module.RunModule(target, flags, scheme)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", flags["module"], registeredModules)
			os.Exit(1)
		}
	}

}
