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

	modules "github.com/cotsom/CloudExec/internal/modules/gitlab"
	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type GitlabModule interface {
	RunModule(target string, flags map[string]string, scheme string)
}

var gitlabdModules = map[string]GitlabModule{
	"loginbypass": modules.Loginbypass{},
	"accesslvl":   modules.Accesslvl{},
	"clone":       modules.Clone{},
	// Add another modules here
}

func init() {
	rootCmd.AddCommand(gitlabCmd)

	gitlabCmd.Flags().IntP("threads", "t", 100, "threads lol")
	gitlabCmd.Flags().StringP("port", "", "", "port lol")
	gitlabCmd.Flags().StringP("user", "u", "", "user lol")
	gitlabCmd.Flags().StringP("password", "p", "", "password lol")
	gitlabCmd.Flags().StringP("inputlist", "i", "", "password inputlist")
	gitlabCmd.Flags().StringP("module", "M", "", "Choose one of module")
	gitlabCmd.Flags().StringP("token", "", "", "set token")
	gitlabCmd.Flags().StringP("timeout", "", "", "Count of seconds for waiting http response")
}

// gitlabCmd represents the gitlab command
var gitlabCmd = &cobra.Command{
	Use:   "gitlab",
	Short: "discover & exploit Gitlab",
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
			go checkGitlab(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
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
		if module, exists := gitlabdModules[flags["module"]]; exists {
			module.RunModule(target, flags, scheme)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", flags["module"], gitlabdModules)
			os.Exit(1)
		}
	}

}
