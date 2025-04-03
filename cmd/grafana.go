/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	modules "github.com/cotsom/CloudExec/internal/modules/grafana"
	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(grafanaCmd)

	grafanaCmd.Flags().IntP("threads", "t", 100, "Number of threads for scan")
	grafanaCmd.Flags().StringP("port", "", "", "Grafana port")
	grafanaCmd.Flags().StringP("user", "u", "", "Grafana user")
	grafanaCmd.Flags().StringP("password", "p", "", "Grafana password")
	grafanaCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	grafanaCmd.Flags().StringP("module", "M", "", "Choose module")
}

type GrafanaModule interface {
	RunModule(target string, flags map[string]string)
}

var grafanadModules = map[string]GrafanaModule{
	"datasource": modules.Datasource{},
	"defcreds":   modules.Defcreds{},
	// Add another modules here
}

// grafanaCmd represents the grafana command
var grafanaCmd = &cobra.Command{
	Use:   "grafana host/subnetwork/input-list",
	Short: "discover & exploit Grafana",
	Long: `Mode for discover & exploit Grafana
Will scan and highlight all found hosts with grafana service.

Modules:
* datasources - displays a list of all available sources for the specified account
* defcreds - try to authenticate with popular creds`,
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
		// fmt.Println(targets)

		//MAIN LOGIC
		var wg sync.WaitGroup
		var sem chan struct{}

		//set threads
		threads, err := strconv.Atoi(flags["threads"])
		if err != nil {
			fmt.Println("You have to set correct number of threads")
			os.Exit(0)
		}
		sem = make(chan struct{}, threads)

		progress := 0
		for i, target := range targets {
			wg.Add(1)
			sem <- struct{}{}
			go checkGrafana(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func checkGrafana(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	if flags["port"] == "" {
		flags["port"] = "3000"
	}

	creds := fmt.Sprintf("%s:%s", flags["user"], flags["password"])

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	//check grafana port
	url := fmt.Sprintf("http://%s:%s", target, flags["port"])
	// fmt.Println(url)

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

	url = fmt.Sprintf("http://%s@%s:%s/api/datasources", creds, target, flags["port"])
	response, err = utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		if flags["user"] == "" && flags["password"] == "" {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Grafana with public dashboards! (%s)\n", utils.ClearLine, target, flags["port"], creds))
		}
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Grafana! (%s)\n", utils.ClearLine, target, flags["port"], creds))
	} else {
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Grafana\n", utils.ClearLine, target, flags["port"]))
	}

	if flags["module"] != "" {
		if module, exists := grafanadModules[flags["module"]]; exists {
			module.RunModule(target, flags)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", module, grafanadModules)
			os.Exit(1)
		}
	}

}
