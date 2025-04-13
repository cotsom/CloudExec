# CloudExec - Cloud Execution Tool
This utility is designed to scan, detect, exploit vulnerabilities and services misconfigurations in cloud and dev linux infrastructures.

Here are mods for scan some services
- Grafana
- Gitlab
- Postgres
- Zookeeper
- Kafka
- Registry
- Kube

### Legend
Blue highlighting - Target found `[*] 192.168.1.1 - Service`

Green highlighting - Target found and access granted  `[+] 192.168.1.1 - Service`

Yellow highlighted `Pwned` - You can execute code (RCE)


# Modes
### Grafana
This mode  is designed to discover & exploit Grafana. It will scan and highlight all found hosts with grafana service

Usage: `clx grafana <ip/network/hostname>`

Modules:

modules usage: `clx grafana <ip/network/hostname> -M moduleName`
* **datasources** - Displays a list of all available sources for the specified account. By querying the data sources, you can retrieve the data stored in them (require `-u` and `-p` flags for authenticate)
* **defcreds** - Try to authenticate with popular creds

# Install
Go 1.23+

`$ go install github.com/cotsom/CloudExec@latest`

# Writing module
You can add your own module using e.g. `cobra-cli add <commandName>` command.  This will create a new file in the `cmd/` directory containing the code template for the new command.

### Template for new mode:
*In the Run function you can implement your own logic for receiving and parsing flags, targets and scans*
```go
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

	modules "github.com/cotsom/CloudExec/internal/modules/test"
	utils "github.com/cotsom/CloudExec/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().IntP("threads", "t", 100, "Number of threads for scan")
	testCmd.Flags().StringP("port", "", "", "Test port")
	testCmd.Flags().StringP("user", "u", "", "Test user")
	testCmd.Flags().StringP("password", "p", "", "Test password")
	testCmd.Flags().StringP("inputlist", "i", "", "Input from list of hosts")
	testCmd.Flags().StringP("module", "M", "", "Choose module")
}

type TestModule interface {
	RunModule(target string, flags map[string]string)
}

var TestdModules = map[string]TestModule{
	"testModule": modules.Test{},
	// Add another modules here
}

// testCmd represents the Test command
var testCmd = &cobra.Command{
	Use:   "usage",
	Short: "short description",
	Long: `long description`,
	Run: func(cmd *cobra.Command, args []string) {
		//Parse flags
		flags := make(map[string]string)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			flags[f.Name] = f.Value.String()
		})

		////Parse targets
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

		//Start check function on all targets with goroutines
		progress := 0
		for i, target := range targets {
			wg.Add(1)
			sem <- struct{}{}
			go checkTest(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func checkTest(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	//defer with free semaphor
	defer func() {
		<-sem
		wg.Done()
	}()


	//========================
	//Some scan logic here   |
	//========================


	//Execute defined module
	if flags["module"] != "" {
		if module, exists := TestdModules[flags["module"]]; exists {
			module.RunModule(target, flags)
		} else {
			fmt.Printf("Module \"%s\" not found. Available modules: %v\n", module, TestdModules)
			os.Exit(1)
		}
	}

}
```
