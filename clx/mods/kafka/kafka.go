package main

import (
	modules "clx/mods/kafka/modules"
	utils "clx/utils"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// mode type for plugin's symbol
type mode string

type Module interface {
	RunModule(target string, flags map[string]string, conn *kafka.Conn, dialer *kafka.Dialer)
}

var registeredModules = map[string]Module{
	"topics": modules.Topics{},
	// Add another modules here
}

func getFlags(args []string) map[string]string {
	requiredParams := map[string]string{
		"-M":          "module",
		"-t":          "threads",
		"--port":      "port",
		"-u":          "user",
		"-p":          "password",
		"-iL":         "inputlist",
		"--topic":     "topic",
		"--mechanism": "mechanism",
	}

	flags := make(map[string]string)

	for key, name := range requiredParams {
		value, err := utils.GetParam(args, key)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		flags[name] = value
	}

	return flags
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
		if module, exists := registeredModules[flags["module"]]; exists {
			module.RunModule(target, flags, conn, dialer)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", flags["module"], registeredModules)
			os.Exit(1)
		}
	}

}

// Main func
func (m mode) Run(args []string) {
	if len(args) < 1 {
		fmt.Println("Enter host or subnetwork")
		return
	}

	flags := getFlags(args)
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
}

// exporting symbol "Mode"
var Mode mode
