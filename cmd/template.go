package cmd

import (
	"github.com/cotsom/CloudExec/internal/resource"
	clickResources "github.com/cotsom/CloudExec/internal/resource/clickhouse"

	"github.com/spf13/cobra"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type TemplateCmd struct {
	resource.Command

	// Redefining
	Opts clickResources.ClickhouseOptions
}

func NewTemplateCmd(opts clickResources.ClickhouseOptions) *TemplateCmd {
	c := &TemplateCmd{
		Opts: opts,
		// ...
	}

	// Sets child `Check` function realization for parent interface
	c.Command.CommandIface = c

	return c
}

// Default command method with main functionality
func (c *TemplateCmd) Check(target string) error {
	return nil
}

func NewCmdTemplate() *cobra.Command {
	o := clickResources.ClickhouseOptions{}

	c := NewTemplateCmd(o)

	cmd := &cobra.Command{
		Use:   "clickhouse",
		Short: "discover Template Service",
		Run:   c.Run,
	}

	// Set default opts from parent
	c.SetDefaultOptions(cmd)

	// Reset default attributes
	cmd.Flags().IntVarP(&c.Opts.Port, "port", "P", 1337, "")

	// Set not default options

	// Modules
	// c.RegisterModule(modules.NewClickhouseBruteModule(c.Opts, c.Logger))

	return cmd
}

func init() {
	// To add Command:
	// rootCmd.AddCommand(NewCmdClickhouse())
}
