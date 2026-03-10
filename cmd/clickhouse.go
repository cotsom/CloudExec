package cmd

import (
	"database/sql"
	"fmt"

	modules "github.com/cotsom/CloudExec/internal/modules/clickhouse"
	"github.com/cotsom/CloudExec/internal/resource"
	clickResources "github.com/cotsom/CloudExec/internal/resource/clickhouse"
	"github.com/cotsom/CloudExec/internal/utils/sqlquery"

	"github.com/spf13/cobra"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickhouseCmd struct {
	resource.Command

	// Redefining
	Opts clickResources.ClickhouseOptions

	// TODO: Mutex to fix print race
}

func NewClickhouseCmd(opts clickResources.ClickhouseOptions) *ClickhouseCmd {
	c := &ClickhouseCmd{
		Opts: opts,
		// ...
	}

	// Sets child `Check` function realization for parent interface
	c.Command.CommandIface = c

	return c
}

// Default command method with main functionality
func (c *ClickhouseCmd) Check(target string) error {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=%ds&read_timeout=%ds",
		c.Opts.Username,
		c.Opts.Password,
		target,
		c.Opts.Port,
		c.Opts.Database,

		c.Opts.Timeout,
		c.Opts.Timeout,
	)

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		c.Logger.Fatal(err.Error())
		return sqlquery.ConnectionFailure
	}
	defer db.Close()

	conn := sqlquery.NewExecutor(db)
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		switch err {
		case sqlquery.ConnectionFailure:
			return err
		case sqlquery.AuthFailure:
			if c.Opts.Username == "" {
				c.Logger.Info(fmt.Sprintf("Cickhouse: %s", target))
			} else {
				c.Logger.Error(fmt.Sprintf("Cickhouse: %s - %s:%s", target, c.Opts.Username, c.Opts.Password))
			}
		case sqlquery.DatabaseFailure:
			c.Logger.Error(fmt.Sprintf("Cickhouse: %s - %s:%s\tDatabase doesn't exist: %s", target, c.Opts.Username, c.Opts.Password, c.Opts.Database))
		default:
			c.Logger.Fatal(err.Error())
			return err
		}
		return nil
	}

	c.Logger.Found(fmt.Sprintf("Cickhouse: %s - %s:%s", target, c.Opts.Username, c.Opts.Password))

	if c.Opts.Query == "" {
		return nil
	}

	rows, err := conn.ExecuteQuery(c.Opts.Query)
	if err != nil {
		c.Logger.Fatal(err.Error())
		return nil
	}
	defer rows.Close()

	fmt.Println(conn.PrintableRows(rows))
	return nil
}

func NewCmdClickhouse() *cobra.Command {
	o := clickResources.ClickhouseOptions{}

	c := NewClickhouseCmd(o)

	cmd := &cobra.Command{
		Use:   "clickhouse",
		Short: "discover Clickhouse",
		Run:   c.Run,
	}

	// Set default opts from parent
	c.SetDefaultOptions(cmd)

	// Reset default attributes
	cmd.Flags().IntVarP(&c.Opts.Port, "port", "P", 9000, "")

	// Set not default options
	cmd.Flags().StringVarP(&c.Opts.Username, "username", "u", "", "")
	cmd.Flags().StringVarP(&c.Opts.Password, "password", "p", "", "")
	cmd.Flags().StringVarP(&c.Opts.Database, "database", "d", "default", "")
	cmd.Flags().IntVarP(&c.Opts.Timeout, "timeout", "", 5, "")
	cmd.Flags().StringVarP(&c.Opts.Query, "query", "q", "", "SQL query to execute after auth")

	// Modules
	c.RegisterModule(modules.NewClickhouseBruteModule(c.Opts, c.Logger))

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCmdClickhouse())
}
