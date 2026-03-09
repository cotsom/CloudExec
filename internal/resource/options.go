package resource

import (
	"errors"
	"sync"

	"github.com/cotsom/CloudExec/internal/utils"
	"github.com/spf13/cobra"
)

type OptionsIface interface {
	Check(target string)
}

type Options struct {
	OptionsIface
	Inputlist string
	Port      int

	ListModules bool
	Module      string

	Threads int

	Logger Logger
}

func (o *Options) SetDefaultOptions(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&o.Threads, "threads", "t", 100, "Number of threads for scan")

	cmd.Flags().StringVarP(&o.Inputlist, "inputlist", "i", "", "Input from file with hosts")

	cmd.Flags().BoolVarP(&o.ListModules, "list-modules", "L", false, "Lists modules")
	cmd.Flags().StringVarP(&o.Module, "module", "M", "", "Choose module")
}

func (o *Options) GetTargets(args []string) ([]string, error) {
	var targets []string

	if (len(args) < 1) && (o.Inputlist == "") {
		return nil, errors.New("Enter: [host / subnetwork / input list (-i)]")
	}

	if o.Inputlist != "" {
		targets = utils.ParseTargetsFromList(o.Inputlist)
	} else {
		targets = utils.ParseTargets(args[0])
	}
	return targets, nil
}

func (o *Options) Run(cmd *cobra.Command, args []string) {
	// TODO: list modules

	// Parse targets
	targets, err := o.GetTargets(args)
	if err != nil {
		o.Logger.Fatal(err.Error())
		cmd.Help()
		return
	}

	// Creates
	var wg sync.WaitGroup
	var sem chan struct{} = make(chan struct{}, o.Threads)

	// Start check function on all targets with goroutines
	// progress := 0
	// for i, target := range targets {
	for _, target := range targets {

		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			o.Check(target)
		}()
		// utils.ProgressBar(len(targets), i+1, &progress)
	}
	wg.Wait()
}
