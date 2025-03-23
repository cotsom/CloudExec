/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	utils "github.com/cotsom/CloudExec/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(postgresCmd)

	postgresCmd.Flags().IntP("threads", "t", 100, "threads")
	postgresCmd.Flags().StringP("port", "", "", "port")
	postgresCmd.Flags().StringP("user", "u", "", "user")
	postgresCmd.Flags().StringP("password", "p", "", "password")
	postgresCmd.Flags().StringP("inputlist", "i", "", "inputlist")
	postgresCmd.Flags().StringP("module", "M", "", "Choose one of module")
}

// postgresCmd represents the postgres command
var postgresCmd = &cobra.Command{
	Use:   "postgres",
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

		if (len(args) < 1) && (flags["inputlist"] != "") {
			fmt.Println("Enter host / subnetwork / input list")
			return
		}

		targets := utils.GetTargets(flags, args)

		// var targets []string
		// if flags["inputlist"] != "" {
		// 	targets = utils.ParseTargetsFromList(flags["inputlist"])
		// } else {
		// 	targets = utils.ParseTargets(args[0])
		// }
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
			go checkPostgres(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func checkPostgres(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
}
