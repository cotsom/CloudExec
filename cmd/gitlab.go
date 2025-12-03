/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
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
	"runnerrce":   modules.RunnerRce{},
	// Add another modules here
}

func init() {
	rootCmd.AddCommand(gitlabCmd)

	gitlabCmd.Flags().IntP("threads", "t", 100, "Number of threads for scan")
	gitlabCmd.Flags().StringP("port", "", "", "Gitlab port")
	gitlabCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	gitlabCmd.Flags().StringP("module", "M", "", "Choose module")
	gitlabCmd.Flags().StringP("token", "", "", "Set auth token")
	gitlabCmd.Flags().StringP("timeout", "", "", "Count of seconds for waiting http response")
	gitlabCmd.Flags().StringP("pjid", "", "", "Project id")
	gitlabCmd.Flags().BoolP("public", "", false, "Use public access")
	gitlabCmd.Flags().String("revshell", "", "Flag for runnerrce module, provide ip:port")
	gitlabCmd.Flags().StringP("exec", "x", "", "Command which will executed via script operator in pipeline")
}

// gitlabCmd represents the gitlab command
var gitlabCmd = &cobra.Command{
	Use:   "gitlab host/subnetwork/input-list",
	Short: "discover & exploit Gitlab",
	Long: `Mode for discover & exploit Gitlab
Will scan and highlight all found hosts with gitlab service.

Modules:
* loginbypass - try endpoints to bypass the login page and get public projects
* accesslvl (Require --token flag) - check personal and group access token rights of all available projects
* clone - clone all available repositories. Add --public flag if u want to clone public repositories
* runnerrce - (Require --token, --revshell or -x, and pjid flags), from accesslvl module u can get id project when u have lvl > 30, for -x flag maybe need more timeout`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := make(map[string]string)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Value.Type() == "bool" {
				if f.Changed {
					flags[f.Name] = f.Value.String()
				}
			} else {
				flags[f.Name] = f.Value.String()
			}
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

	gitlabRoute := "users/sign_in"

	port := flags["port"]
	if port == "" {
		port = "80"
	}

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	// first https, next http
	schemes := []string{"https", "http"}

	var detectedScheme string

	for _, scheme := range schemes {
		url := fmt.Sprintf("%s://%s:%s/%s", scheme, target, port, gitlabRoute)

		response, err := utils.HttpRequest(url, http.MethodGet, nil, client)
		if err != nil {
			//next scheme
			continue
		}

		func() {
			defer response.Body.Close()
			respBody, err := io.ReadAll(response.Body)
			if err != nil {
				return
			}
			if response.StatusCode < 200 || response.StatusCode >= 400 {
				return
			}
			bodyLower := strings.ToLower(string(respBody))
			if !strings.Contains(bodyLower, "gitlab") {
				return
			}
			detectedScheme = scheme
		}()

		if detectedScheme != "" {
			break
		}
	}

	if detectedScheme == "" {
		return
	}

	utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Gitlab\n", utils.ClearLine, target, port))

	if flags["module"] != "" {
		if module, exists := gitlabdModules[flags["module"]]; exists {
			module.RunModule(target, flags, detectedScheme)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", flags["module"], gitlabdModules)
			os.Exit(1)
		}
	}
}
