package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"regexp"

	utils "github.com/cotsom/CloudExec/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// redisCmd represents the etcd command
var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "discover redis",
	Long: `Mode for discover redis
Will scan and highlight all found hosts with redis.

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
			go redisMode(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(redisCmd)

	redisCmd.Flags().IntP("threads", "t", 100, "threads")
	redisCmd.Flags().StringP("port", "", "", "etcd port")
	redisCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	redisCmd.Flags().StringP("module", "M", "", "Choose module")
	redisCmd.Flags().StringP("timeout", "", "2", "Count of seconds for waiting http response")
	redisCmd.Flags().Bool("keycount", false, "Check count keys, maybe need more timeout")
	redisCmd.Flags().StringP("password", "p", "", "Password or wordlist")

}

func redisMode(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	var iswordlist bool
	var wordlist []string
	//var successLogon bool
	defer func() {
		<-sem
		wg.Done()
	}()
	port := flags["port"]
	if port == "" {
		port = "6379"
	}
	keycountNeed, _ := strconv.ParseBool(flags["keycount"])
	timeout, err := strconv.Atoi(flags["timeout"])
	if err != nil {
		utils.Colorize(utils.ColorRed, "Invalid timeout value")
		return
	}
	if flags["password"] != "" {
		//try openfile
		bytes, err := os.ReadFile(flags["password"])
		if err == nil {
			iswordlist = true
			wordlist = strings.Split(string(bytes), "\n")
		}

	}
	if !detectRedis(target, port, timeout) {
		return
	}
	if iswordlist {
		for _, passwd := range wordlist {
			checkRedis(passwd, timeout, target, port, keycountNeed)
		}
		return
	}

	checkRedis(flags["password"], timeout, target, port, keycountNeed)

}

func checkRedis(password string, timeout int, target string, port string, keycountNeed bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", target, port),
		Password: password,
		DB:       0,
	})
	val, err := rdb.Info(ctx, "keyspace").Result()
	if err != nil {

		if strings.Contains(err.Error(), "WRONGPASS") {
			utils.Colorize(utils.ColorRed, fmt.Sprintf("[-] %s - Redis | Password: %s", target, password))
			return
		}

	}
	if strings.Contains(val, "# Keyspace") {
		if password != "" {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - Redis | Password: %s", target, password))

		} else {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - Redis", target))
		}
	}
	if keycountNeed {
		var count int32
		re := regexp.MustCompile("keys=([^,]*)")
		matches := re.FindAllString(val, -1)
		for _, match := range matches {
			c, err := strconv.Atoi(strings.Split(match, "=")[1])
			if err == nil {
				count += int32(c)
			}
		}
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - Redis, %d keys", target, count))
	}
}

func detectRedis(target string, port string, timeout int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", target, port),
		Password: "",
		DB:       0,
	})
	_, err := rdb.Info(ctx, "keyspace").Result()
	if strings.Contains(err.Error(), "NOAUTH") {
		utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - Redis", target))
		return true
	}
	return false
}
