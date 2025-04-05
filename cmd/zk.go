/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	utils "github.com/cotsom/CloudExec/internal/utils"
	"github.com/go-zookeeper/zk"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(zkCmd)

	zkCmd.Flags().IntP("threads", "t", 100, "threads")
	zkCmd.Flags().StringP("port", "", "", "Zookeeper port")
	zkCmd.Flags().StringP("user", "u", "", "Zookeeper user")
	zkCmd.Flags().StringP("password", "p", "", "Zookeeper password")
	zkCmd.Flags().StringP("inputlist", "i", "", "inputlist")
	zkCmd.Flags().StringP("module", "M", "", "Choose one of module")
	zkCmd.Flags().StringP("list", "", "", "List znodes")
	zkCmd.Flags().StringP("get", "", "", "Get znode")
}

// zkCmd represents the zk command
var zkCmd = &cobra.Command{
	Use:   "zk",
	Short: "A brief description of your command",
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
			go checkZookeeper(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func checkZookeeper(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	port := flags["port"]
	if port == "" {
		port = "2181"
	}

	c, _, err := zk.Connect([]string{fmt.Sprintf("%s:%s", target, port)}, time.Second) //*10)
	if err != nil {
		fmt.Println(err)
		if !strings.Contains(err.Error(), "connection refused") {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Zookeeper\n", utils.ClearLine, target, port))
		}
		return
	}

	err = c.AddAuth("digest", []byte(fmt.Sprintf("%s:%s", flags["user"], flags["password"])))
	fmt.Println(err)
	if err != nil {
		fmt.Println(err)
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Zookeeper\n", utils.ClearLine, target, port))
		return
	}

	utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Zookeeper\n", utils.ClearLine, target, port))

	// switch znodeAction := flags

	if flags["list"] != "" {
		children, _, _, err := c.ChildrenW(flags["list"])
		if err != nil {
			utils.Colorize(utils.ColorRed, err.Error())
		}
		utils.Colorize(utils.ColorYellow, fmt.Sprintf("%+v", children))
		if len(children) == 0 {
			utils.Colorize(utils.ColorYellow, "Try to use --get flag")
		}
	}

	if flags["get"] != "" {
		out, _, err := c.Get(flags["get"])
		if err != nil {
			utils.Colorize(utils.ColorRed, err.Error())
		}
		utils.Colorize(utils.ColorYellow, string(out))
		if string(out) == "" {
			utils.Colorize(utils.ColorYellow, "Try to use --list flag")
		}
	}

}
