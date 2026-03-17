package resource

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cotsom/CloudExec/internal/utils"
	"github.com/spf13/cobra"
)

// TODO: think about no interface
type CommandIface interface {
	Check(target string) error
}

type Command struct {
	CommandIface

	Name string
	Opts Options

	Logger  *Logger
	Modules map[string]ModuleIface
}

func (c *Command) SetDefaultOptions(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&c.Opts.Threads, "threads", "t", 100, "Number of threads for scan")

	cmd.Flags().StringVarP(&c.Opts.Inputlist, "inputlist", "i", "", "Input from file with hosts")

	cmd.Flags().BoolVarP(&c.Opts.ListModules, "list-modules", "L", false, "Lists modules")
	cmd.Flags().StringVarP(&c.Opts.Module, "module", "M", "", "Choose module")

	// For command name
	c.Name = cmd.Use
}

func (c *Command) RegisterModule(module ModuleIface) {
	if c.Modules == nil {
		c.Modules = make(map[string]ModuleIface)
	}

	c.Modules[module.GetName()] = module
}

func (c *Command) GetTargets(args []string) ([]string, error) {
	var targets []string

	if (len(args) < 1) && (c.Opts.Inputlist == "") {
		return nil, errors.New("Enter: [host / subnetwork / input list (-i)]")
	}

	if c.Opts.Inputlist != "" {
		targets = utils.ParseTargetsFromList(c.Opts.Inputlist)
	} else {
		targets = utils.ParseTargets(args[0])
	}
	return targets, nil
}

func (c *Command) Run(cmd *cobra.Command, args []string) {
	defer c.Logger.DeferPrint()

	if c.Opts.ListModules {
		if len(c.Modules) == 0 {
			c.Logger.Raw(
				fmt.Sprintf("No such modules for %s :(", c.Name),
			)
			cmd.Help()
			return
		}

		c.Logger.Raw(
			fmt.Sprintf("Modules list for %s:", c.Name),
		)
		for _, module := range c.Modules {
			c.Logger.List(
				fmt.Sprintf("%s - %s", module.GetName(), module.GetDescription()),
			)
		}
		return
	}

	// Parse targets
	targets, err := c.GetTargets(args)
	if err != nil {
		c.Logger.Fatal(err.Error())
		cmd.Help()
		return
	}
	// Creates
	var wg sync.WaitGroup
	var sem chan struct{} = make(chan struct{}, c.Opts.Threads)

	// TODO: progress bar
	// progress := 0
	// for i, target := range targets {
	for _, target := range targets {

		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer func() {
				<-sem
				wg.Done()
				c.Logger.DeferPrint()
			}()
			// Running check of default connection
			err := c.Check(target)
			// If target is alive
			if err != nil || c.Opts.Module == "" {
				return
			}
			// Run module
			module, ok := c.Modules[c.Opts.Module]
			if !ok {
				c.Logger.Fatal(
					fmt.Sprintf("No such nodule %s\n Try -L flag to list all flags", c.Opts.Module),
				)
				return
			}
			module.Run(target)
		}()
		// utils.ProgressBar(len(targets), i+1, &progress)
	}
	wg.Wait()
}
