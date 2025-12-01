package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	modules "github.com/cotsom/CloudExec/internal/modules/consul"
	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type agentSelf struct {
	DebugConfig struct {
		ACLsEnabled              *bool `json:"ACLsEnabled"`
		EnableRemoteScriptChecks *bool `json:"EnableRemoteScriptChecks"`
	} `json:"DebugConfig"`
}

func init() {
	rootCmd.AddCommand(consulCmd)

	consulCmd.Flags().IntP("threads", "t", 100, "Number of threads for scan")
	consulCmd.Flags().StringP("port", "", "", "Consul port")
	consulCmd.Flags().StringP("user", "u", "", "Consul user")
	consulCmd.Flags().StringP("password", "p", "", "Consul password")
	consulCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	consulCmd.Flags().StringP("module", "M", "", "Choose module")
	consulCmd.Flags().StringP("ssrf-target", "", "", "target for SSRF module")
	consulCmd.Flags().StringP("ssrf-network", "", "", "network for scan in SSRF module (e.g 192.168.1.0/24)")
	consulCmd.Flags().StringP("ssrf-port", "", "", "port for ssrf module")
	consulCmd.Flags().StringP("timeout", "", "", "Count of seconds for waiting http response")
	consulCmd.Flags().StringP("exec", "x", "", "execute a command with RCE module")
}

type ConsulModule interface {
	RunModule(target string, flags map[string]string, scheme string)
}

var consulModules = map[string]ConsulModule{
	"ssrf": modules.Ssrf{},
	"rce":  modules.Rce{},
	// Add another modules here
}

// consulCmd represents the consul command
var consulCmd = &cobra.Command{
	Use:   "consul",
	Short: "discover & exploit Consul",
	Long: `Mode for discover & exploit Consul
Will scan and highlight all found hosts with consul service. "Pwned!" suggets rce availabe

Modules:
* ssrf - send http request to one target or all targets in network on behalf of Consul
* rce - this module triggers with -x flag. If you see "Pwned!" is suggets that rce availabe and you have availability to execute a command`,
	Run: func(cmd *cobra.Command, args []string) {
		//Parse flags
		flags := make(map[string]string)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			flags[f.Name] = f.Value.String()
		})

		////Parse targets
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

		//Start check function on all targets with goroutines
		progress := 0
		for i, target := range targets {
			wg.Add(1)
			sem <- struct{}{}
			go checkConsul(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func checkConsul(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	//defer with free semaphor
	defer func() {
		<-sem
		wg.Done()
	}()

	if flags["port"] == "" {
		flags["port"] = "8500"
	}
	if flags["timeout"] == "" {
		flags["timeout"] = "3"
	}

	timeout, _ := strconv.Atoi(flags["timeout"])
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	isConsul, aclEnabled, rceEnabled, scheme := isConsul(target, flags["port"], client, flags)

	// fmt.Println(isConsul, aclEnabled, rceEnabled, scheme)

	if !isConsul {
		return
	}

	if aclEnabled && rceEnabled {
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Consul %sPwned!%s", utils.ClearLine, target, flags["port"], utils.ColorYellow, utils.ColorReset))
	} else if aclEnabled {
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Consul\n", utils.ClearLine, target, flags["port"]))
	} else {
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Consul\n", utils.ClearLine, target, flags["port"]))
	}

	//Execute defined module
	if flags["module"] != "" {
		if module, exists := consulModules[flags["module"]]; exists {
			module.RunModule(target, flags, scheme)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", module, consulModules)
			os.Exit(1)
		}
	} else if flags["exec"] != "" {
		consulModules["rce"].RunModule(target, flags, scheme)
	}

}

func isConsul(target string, port string, client http.Client, flags map[string]string) (bool, bool, bool, string) {
	scheme := "http"
	consulRoute := "v1/agent/self"

	// Make http req
	url := fmt.Sprintf("http://%s:%s/%s", target, flags["port"], consulRoute)
	response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)

	if err != nil {
		// utils.Colorize(utils.ColorRed, fmt.Sprintf("%s[!] %s:%s - %s\n", utils.ClearLine, target, flags["port"], err))
		return false, false, false, ""
	}
	defer response.Body.Close()
	respBody, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
	}

	// Make https req
	if strings.Contains(string(respBody), "HTTP request was sent to HTTPS port") {
		url = fmt.Sprintf("https://%s:%s/%s", target, flags["port"], consulRoute)

		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			// utils.Colorize(utils.ColorRed, fmt.Sprintf("%s[!] %s:%s - %s\n", utils.ClearLine, target, flags["port"], err))
			return false, false, false, ""
		}
		defer response.Body.Close()
		respBody, err = ioutil.ReadAll(response.Body)

		if err != nil {
			utils.Colorize(utils.ColorRed, fmt.Sprintf("%s[!] %s:%s - %s\n", utils.ClearLine, target, flags["port"], err))
			return false, false, false, ""
		}
		scheme = "https"
	}

	switch response.StatusCode {
	case 200:
		var config agentSelf

		err = json.Unmarshal(respBody, &config)
		if err != nil {
			// fmt.Println("Error unmarshalling JSON:", err)
			return false, false, false, ""
			// return true, false, false, scheme
		}

		aclsCheck := !*config.DebugConfig.ACLsEnabled
		rceCheck := config.DebugConfig.EnableRemoteScriptChecks

		return true, aclsCheck, *rceCheck, scheme

	case 403:
		if strings.Contains(string(respBody), "token lacks permission") {
			return true, false, false, scheme
		}
	default:
		if strings.Contains(string(respBody), "consul") {
			return true, false, false, scheme
		}
	}

	return false, false, false, ""
}
