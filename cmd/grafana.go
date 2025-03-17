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
	Use:   "grafana",
	Short: "discover & exploit Grafana",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := make(map[string]string)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			flags[f.Name] = f.Value.String()
		})

		if (len(args) < 1) && (flags["inputlist"] != "") {
			fmt.Println("Enter host / subnetwork / input list")
			return
		}

		var targets []string
		if flags["inputlist"] != "" {
			targets = utils.ParseTargetsFromList(flags["inputlist"])
		} else {
			targets = utils.ParseTargets(args[0])
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

// var (
// 	// Used for flags.
// 	threads   int
// 	port      string
// 	user      string
// 	password  string
// 	inputlist string
// 	module    string
// )

func init() {
	rootCmd.AddCommand(grafanaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// grafanaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// grafanaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	grafanaCmd.Flags().IntP("threads", "t", 100, "threads lol")
	grafanaCmd.Flags().StringP("port", "", "", "port lol")
	grafanaCmd.Flags().StringP("user", "u", "", "user lol")
	grafanaCmd.Flags().StringP("password", "p", "", "password lol")
	grafanaCmd.Flags().StringP("inputlist", "i", "", "inputlist")
	grafanaCmd.Flags().StringP("module", "M", "", "Choose one of module")
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
