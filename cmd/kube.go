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

	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// kubeCmd represents the kube command
var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "discover & exploit Kubernetes",
	Long: `Mode for discover & exploit Kubernetes
Will scan and highlight all found hosts with kubernetes & kubelets.

Modules:
-`,
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
			go checkKube(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(kubeCmd)

	kubeCmd.Flags().IntP("threads", "t", 100, "threads")
	kubeCmd.Flags().StringP("port", "", "", "kubernetes port")
	kubeCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	kubeCmd.Flags().StringP("module", "M", "", "Choose module")
	kubeCmd.Flags().StringP("timeout", "", "", "Count of seconds for waiting http response")
}

func checkKube(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	// fmt.Println(target)
	defer func() {
		<-sem
		wg.Done()
	}()
	ports := map[string][]string{"kubeapi": {"6443"}, "kubelet": {"10250"}}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	//check kubeapi
	for _, port := range ports["kubeapi"] {
		url := fmt.Sprintf("https://%s:%s", target, port)

		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		defer response.Body.Close()

		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		if strings.Contains(string(respBody), "\"apiVersion\"") {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - kube Api", target))
		}
	}

	//check kubelet
	for _, port := range ports["kubelet"] {
		url := fmt.Sprintf("https://%s:%s/pods", target, port)
		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			// fmt.Println(err)
			continue
		}

		defer response.Body.Close()

		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		if strings.Contains(string(respBody), "Unauthorized") {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - kubelet", target))
		} else {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - kubelet UNAUTH!", target))
		}
	}
}
