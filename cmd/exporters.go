/*
Copyright © 2026 oyama forked cotsom
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

	"github.com/cotsom/CloudExec/internal/secretsengine"
	"github.com/cotsom/CloudExec/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	RelayFlag          = false
	RelayIP            = ""
	ExportersPortBegin = 9180
	ExportersPortEnd   = 9188
)

var RelayEndpoints = map[string]string{
	"postgres":      "/probe?target=",
	"pgbouncer":     "/probe?target=",
	"proxmox":       "/pve?target=",
	"redis":         "/scrape?target=",
	"elasticsearch": "/probe?target=",
}

var DebugEndpoints = []string{"/debug/vars",
	"/debug/pprof/cmdline",
	"/debug/pprof",
}

var exportersCmd = &cobra.Command{
	Use:   "exporters",
	Short: "Prometheus exporters Weaknesses",
	Long: `General Exporters Weaknesses:
			- /debug/pprof/
			- /debug/pprof/cmdline/
			- /debug/vars
			- relay attacks`,
	Run: func(cmd *cobra.Command, args []string) {
		secretsengine.LoadRules()
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
		threads, err := strconv.Atoi(flags["threads"])
		if err != nil {
			fmt.Println("You have to set correct number of threads")
			os.Exit(0)
		}
		sem = make(chan struct{}, threads)

		//progress := 0
		// temp disable this progress bar
		// i do only regress
		for _, target := range targets {
			wg.Add(1)
			sem <- struct{}{}
			go checkExporters(target, &wg, sem, threads, flags)
			//utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(exportersCmd)
	exportersCmd.Flags().IntP("threads", "t", 100, "threads")
	exportersCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	exportersCmd.Flags().StringP("module", "M", "", "Choose module")
	exportersCmd.Flags().StringP("timeout", "", "", "Count of seconds for waiting http response")
	exportersCmd.Flags().BoolVarP(&RelayFlag, "relay", "r", false, "Enable relay attack")
}

func checkExporters(target string, wg *sync.WaitGroup, sem chan struct{}, threads int, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	ports := detectExportersPort(target, threads)
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	for port, exporterType := range ports {
		for _, endpoint := range DebugEndpoints {
			url := fmt.Sprintf("http://%s:%d%s", target, port, endpoint)
			response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)

			if err != nil {
				fmt.Println(err)
				continue
			}

			defer response.Body.Close()

			respBody, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("client: could not read response body: %s\n", err)
			}

			if endpoint == "/debug/vars" {
				if strings.Contains(string(respBody), "cmdline") || len(string(respBody)) > 0 {
					utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s:%d%s - found (debug/vars)", target, port, endpoint))
					m, err := utils.UnmarshallJsonString(string(respBody))
					if err != nil {
						continue
					}
					if cmdline, ok := utils.ExportersExtractCmdline(m); ok {
						//utils.Colorize(utils.ColorGreen, fmt.Sprintf("[*] %s:%d%s - cmdline: %s", target, port, endpoint, cmdline))
						regulated := secretsengine.FindSecrets(cmdline)
						if len(regulated) > 0 {
							utils.Colorize(utils.ColorRed, fmt.Sprintf("[*] %s:%s%s - cmdline secret: %s", target, strconv.Itoa(port), endpoint, regulated))
						} else {
							utils.Colorize(utils.ColorGreen, fmt.Sprintf("[*] %s:%s%s - cmdline found, but no secrets", target, strconv.Itoa(port), endpoint))
						}
						break
					}
					continue
				}
			}

			if response.StatusCode == 200 {
				utils.Colorize(utils.ColorGreen, fmt.Sprintf("[*] %s:%d%s - found", target, port, endpoint))
			}
		}

		if RelayFlag {
			checkExportersRelay(target, strconv.Itoa(port), exporterType, flags)
		}
	}
}

func checkExportersRelay(target string, port string, exporterType string,
	flags map[string]string) {

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	relayEndpoint, ok := RelayEndpoints[strings.ToLower(exporterType)]
	if !ok {
		return
	}
	url := fmt.Sprintf("http://%s:%s%s", target, port, relayEndpoint)

	response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		return
	}

	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	if strings.Contains(string(respBody), "target") || len(string(respBody)) > 0 {
		utils.Colorize(utils.ColorYellow, fmt.Sprintf("[*] %s:%s%s - potential relay", target, port, relayEndpoint))
	}

	return

}

func detectExportersPort(target string, threads int) map[int]string {
	ports := make(map[int]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, threads)

	client := http.Client{
		Timeout: 2000 * time.Millisecond,
	}

	for port := ExportersPortBegin; port <= ExportersPortEnd; port++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(p int) {
			defer func() {
				<-sem
				wg.Done()
			}()

			url := fmt.Sprintf("http://%s:%d", target, p)
			response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
			if err != nil {
				return
			}
			defer response.Body.Close()

			if response.StatusCode == 200 {
				exporterType, err := utils.ParseExportersType(response.Body)
				if err != nil {
					return
				}

				utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - detected %s on %d port", target, exporterType, p))

				mu.Lock()
				ports[p] = exporterType
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return ports
}
