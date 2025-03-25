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

	modules "github.com/cotsom/CloudExec/internal/modules/kafka"
	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type KafkaModule interface {
	RunModule(target string, flags map[string]string, conn *kafka.Conn, dialer *kafka.Dialer)
}

var kafkaModules = map[string]KafkaModule{
	"topics": modules.Topics{},
	// Add another modules here
}

// kafkaCmd represents the kafka command
var kafkaCmd = &cobra.Command{
	Use:   "kafka",
	Short: "discover & exploit Kafka",
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
		if flags["threads"] != "" {
			threads, err := strconv.Atoi(flags["threads"])
			if err != nil {
				fmt.Println("You have to set correct number of threads")
				os.Exit(0)
			}
			sem = make(chan struct{}, threads)
		} else {
			sem = make(chan struct{}, 100)
		}

		progress := 0
		for i, target := range targets {
			wg.Add(1)
			sem <- struct{}{}
			go checkKafka(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(kafkaCmd)

	kafkaCmd.Flags().IntP("threads", "t", 100, "threads lol")
	kafkaCmd.Flags().StringP("port", "", "", "port lol")
	kafkaCmd.Flags().StringP("user", "u", "", "user lol")
	kafkaCmd.Flags().StringP("password", "p", "", "password lol")
	kafkaCmd.Flags().StringP("inputlist", "i", "", "password inputlist")
	kafkaCmd.Flags().StringP("module", "M", "", "Choose one of module")
	kafkaCmd.Flags().StringP("mechanism", "", "", "Choose mechanism")
	kafkaCmd.Flags().StringP("topic", "", "", "Choose topic to read")
}

func checkKafka(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	if flags["port"] == "" {
		flags["port"] = "9092"
	}
	broker := fmt.Sprintf("%s:%s", target, flags["port"])

	var conn *kafka.Conn
	var err error
	var dialer *kafka.Dialer

	switch flags["mechanism"] {
	case "SASL_PLAIN":
		mechanism := plain.Mechanism{
			Username: flags["user"],
			Password: flags["password"],
		}

		dialer = &kafka.Dialer{
			Timeout:       1 * time.Second,
			DualStack:     true,
			SASLMechanism: mechanism,
		}

		conn, err = dialer.Dial("tcp", broker)
		if err != nil {
			fmt.Println(err)
			utils.Colorize(utils.ColorRed, fmt.Sprintf("%s[-] %s:%s - Kafka (%s:%s)\n", utils.ClearLine, target, flags["port"], flags["user"], flags["password"]))
			return
		}
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Kafka! (%s:%s)\n", utils.ClearLine, target, flags["port"], flags["user"], flags["password"]))
	default:
		dialer = &kafka.Dialer{
			Timeout: 1 * time.Second,
		}

		conn, err = dialer.Dial("tcp", broker)
		if err != nil {
			fmt.Println(err)
			return
		}
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Kafka\n", utils.ClearLine, target, flags["port"]))
	}
	defer conn.Close()

	// Start module on target
	if flags["module"] != "" {
		if module, exists := kafkaModules[flags["module"]]; exists {
			module.RunModule(target, flags, conn, dialer)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", flags["module"], kafkaModules)
			os.Exit(1)
		}
	}

}
