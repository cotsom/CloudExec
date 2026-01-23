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

func init() {
	rootCmd.AddCommand(kafkaCmd)

	kafkaCmd.Flags().IntP("threads", "t", 100, "Number of threads for scan")
	kafkaCmd.Flags().StringP("port", "", "", "Kafka port")
	kafkaCmd.Flags().StringP("user", "u", "", "Kafka user")
	kafkaCmd.Flags().StringP("password", "p", "", "Kafka password")
	kafkaCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	kafkaCmd.Flags().StringP("module", "M", "", "Choose module")
	kafkaCmd.Flags().StringP("mechanism", "", "", "Kafka authentication mechanism (SASL_PLAINTEXT is available, default is PLAINTEXT)")
	kafkaCmd.Flags().StringP("topic", "", "", "Choose topic to read")
}

type KafkaModule interface {
	RunModule(target string, flags map[string]string, conn *kafka.Conn, dialer *kafka.Dialer)
}

var kafkaModules = map[string]KafkaModule{
	"topics": modules.Topics{},
	// Add another modules here
}

// kafkaCmd represents the kafka command
var kafkaCmd = &cobra.Command{
	Use:   "kafka host/subnetwork/input-list",
	Short: "discover & exploit Kafka",
	Long: `Mode for discover & exploit Kafka
Will scan and highlight all found hosts with kafka.

Modules:
* topics - if the topic flag is set, then the module will read the contents of the selected topic, otherwise it will display all available topics in the broker.`,
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

func checkKafka(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	port, err := utils.SetPort(flags["port"], "9092")
	if err != nil {
		utils.Colorize(utils.ColorRed, err.Error())
		return
	}
	broker := fmt.Sprintf("%s:%s", target, port)

	var conn *kafka.Conn
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
			// fmt.Println(err)
			utils.Colorize(utils.ColorRed, fmt.Sprintf("%s[-] %s:%s - Kafka (%s:%s)\n", utils.ClearLine, target, port, flags["user"], flags["password"]))
			return
		}
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Kafka! (%s:%s)\n", utils.ClearLine, target, port, flags["user"], flags["password"]))
	default:
		dialer = &kafka.Dialer{
			Timeout: 1 * time.Second,
		}

		conn, err = dialer.Dial("tcp", broker)
		if err != nil {
			fmt.Println(err)
			return
		}
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Kafka\n", utils.ClearLine, target, port))
	}
	defer conn.Close()

	// Start module on target
	if flags["module"] != "" {
		if module, exists := kafkaModules[flags["module"]]; exists {
			module.RunModule(target, flags, conn, dialer)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", port, kafkaModules)
			os.Exit(1)
		}
	}

}
