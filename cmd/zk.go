/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
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
	zkCmd.Flags().StringP("port", "", "", "port")
	zkCmd.Flags().StringP("user", "u", "", "user")
	zkCmd.Flags().StringP("password", "p", "", "password")
	zkCmd.Flags().StringP("inputlist", "i", "", "inputlist")
	zkCmd.Flags().StringP("module", "M", "", "Choose one of module")
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
		panic(err)
	}
	children, stat, ch, err := c.ChildrenW("/")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v %+v\n", children, stat)
	e := <-ch
	fmt.Printf("%+v\n", e)
}
