/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	utils "github.com/cotsom/CloudExec/internal/utils"
	clientv2 "go.etcd.io/etcd/client/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// etcdCmd represents the etcd command
var etcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "discover & exploit etcd",
	Long: `Mode for discover & exploit etcd
Will scan and highlight all found hosts with etcd.

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
			go checkEtcd(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(etcdCmd)

	etcdCmd.Flags().IntP("threads", "t", 100, "threads")
	etcdCmd.Flags().StringP("port", "", "", "etcd port")
	etcdCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	etcdCmd.Flags().StringP("module", "M", "", "Choose module")
	etcdCmd.Flags().StringP("timeout", "", "2", "Count of seconds for waiting http response")
	etcdCmd.Flags().Bool("keycount", false, "Check count keys, maybe need more timeout")
}

func checkEtcd(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()
	port := flags["port"]
	if port == "" {
		port = "2379"
	}
	keycountNeed, _ := strconv.ParseBool(flags["keycount"])
	timeout, err := strconv.Atoi(flags["timeout"])
	if err != nil {
		utils.Colorize(utils.ColorRed, "Invalid timeout value")
		return
	}

	//check v2
	cli := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := cli.Get(fmt.Sprintf("http://%s:%s/v2/keys", target, port))
	if err != nil {
		resp, _ := cli.Get(fmt.Sprintf("https://%s:%s/v2/keys", target, port))
		if resp != nil {
			if resp.StatusCode == 200 {
				utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - etcd", target))
				utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - etcd v2", target))
			} else {
				resp, _ := cli.Get(fmt.Sprintf("https://%s:%s/version", target, port))
				body, _ := ioutil.ReadAll(resp.Body)
				if strings.Contains(string(body), "etcd") {
					utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - etcd", target))
				}

			}
		}
	} else {
		if resp != nil {
			if resp.StatusCode == 200 {
				utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - etcd", target))
				utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - etcd v2", target))
			}
			resp, _ := cli.Get(fmt.Sprintf("http://%s:%s/version", target, port))
			body, _ := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(body), "etcd") {
				utils.Colorize(utils.ColorBlue, fmt.Sprintf("[*] %s - etcd", target))
			}
		}
	}
	if keycountNeed {
		cli2, err := clientv2.New(clientv2.Config{
			Endpoints: []string{fmt.Sprintf("http://%s:%s", target, port)},
		})
		if err != nil {
			return
		}
		keysAPI := clientv2.NewKeysAPI(cli2)
		keys2, err := keysAPI.Get(context.Background(), "", &clientv2.GetOptions{})
		if err != nil {
			if !strings.Contains(err.Error(), "connection refused") {
				cli2, _ = clientv2.New(clientv2.Config{
					Endpoints: []string{fmt.Sprintf("https://%s:%s", target, port)},
					Transport: &http.Transport{TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					}},
				})
				keysAPI := clientv2.NewKeysAPI(cli2)
				keys2, err = keysAPI.Get(context.Background(), "", &clientv2.GetOptions{})
				if err == nil && len(keys2.Node.Nodes) > 0 {
					utils.Colorize(utils.ColorGreen, fmt.Sprintf("+] %s - etcd v2, %d keys", target, len(keys2.Node.Nodes)))
				}
			}
		} else {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - etcd v2, %d keys", target, len(keys2.Node.Nodes)))
		}
	}

	//check v3
	cli3, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("http://%s:%s", target, port)},
		DialTimeout: time.Duration(timeout) * time.Second,
		Logger:      zap.NewNop(),
	})
	if err != nil {
		return
	}
	getCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	_, err = cli3.Get(getCtx, "", clientv3.WithPrefix(), clientv3.WithLimit(1))
	if err != nil {
		if strings.Contains(err.Error(), "user name is empty") {
			return
		} else {
			cli3, _ := clientv3.New(clientv3.Config{
				Endpoints:   []string{fmt.Sprintf("https://%s:%s", target, port)},
				DialTimeout: time.Duration(timeout) * time.Second,
				TLS: &tls.Config{
					InsecureSkipVerify: true,
				},
				Logger: zap.NewNop(),
			})
			getCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()
			_, err := cli3.Get(getCtx, "", clientv3.WithPrefix(), clientv3.WithLimit(1))
			if err == nil {
				utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - etcd v3", target))
			} else {
				return
			}
		}

	} else {
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - etcd v3", target))
	}

	if keycountNeed {
		keys, err := cli3.Get(getCtx, "", clientv3.WithPrefix())
		if err != nil {
			cli3, _ := clientv3.New(clientv3.Config{
				Endpoints:   []string{fmt.Sprintf("https://%s:%s", target, port)},
				DialTimeout: time.Duration(timeout) * time.Second,
				TLS: &tls.Config{
					InsecureSkipVerify: true,
				},
				Logger: zap.NewNop(),
			})
			getCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()
			keys, err := cli3.Get(getCtx, "", clientv3.WithPrefix())
			if err == nil {
				utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - etcd v3, %d keys", target, len(keys.Kvs)))
			}

		} else {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - etcd v3, %d keys", target, len(keys.Kvs)))
		}
	}

}
