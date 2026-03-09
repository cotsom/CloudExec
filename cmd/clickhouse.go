package cmd

import (

	//modules "github.com/cotsom/CloudExec/internal/modules/test"

	"database/sql"
	"fmt"

	"github.com/cotsom/CloudExec/internal/resource"
	"github.com/cotsom/CloudExec/internal/utils/sqlquery"

	"github.com/spf13/cobra"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickhouseOptions struct {
	resource.Options

	// Some default values for clickhouse
	Username string
	Password string
	Database string

	Query string

	Timeout int
}

func NewClickhouseOptions() *ClickhouseOptions {
	o := &ClickhouseOptions{
		// ...
	}

	// Sets child `Check` function realization for parent interface
	o.Options.OptionsIface = o

	return o
}

// type ClickModule interface {
// 	RunModule(target string, flags map[string]string)
// }

// var ClickhouseModules = map[string]ClickModule{
// 	"testModule": modules.Test{},
// 	// Add another modules here
// }

func (o *ClickhouseOptions) Check(target string) {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=%ds&read_timeout=%ds",
		o.Username,
		o.Password,
		target,
		o.Port,
		o.Database,

		o.Timeout,
		o.Timeout,
	)

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		o.Logger.Fatal(err.Error())
		return
	}
	defer db.Close()

	conn := sqlquery.NewExecutor(db)
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		switch err {
		case sqlquery.ConnectionFailure:
			// Ignore
		case sqlquery.AuthFailure:
			if o.Username == "" {
				o.Logger.Info(fmt.Sprintf("Cickhouse: %s", target))
			} else {
				o.Logger.Error(fmt.Sprintf("Cickhouse: %s - %s:%s", target, o.Username, o.Password))
			}
		case sqlquery.DatabaseFailure:
			o.Logger.Error(fmt.Sprintf("Cickhouse: %s - %s:%s\tDatabase doesn't exist: %s", target, o.Username, o.Password, o.Database))
		default:
			o.Logger.Fatal(err.Error())
		}
		return
	}

	o.Logger.Found(fmt.Sprintf("Cickhouse: %s - %s:%s", target, o.Username, o.Password))

	if o.Query == "" {
		return
	}

	rows, err := conn.ExecuteQuery(o.Query)
	if err != nil {
		o.Logger.Fatal(err.Error())
		return
	}
	defer rows.Close()

	fmt.Println(conn.PrintableRows(rows))
}

func NewCmdClickhouse() *cobra.Command {
	o := NewClickhouseOptions()
	cmd := &cobra.Command{
		Use:   "clickhouse",
		Short: "discover Clickhouse",
		Long:  `This module discoveres ClickHouse Database`,
		Run:   o.Run,
	}

	o.SetDefaultOptions(cmd)

	// Reset default attributes
	cmd.Flags().IntVarP(&o.Port, "port", "P", 9000, "")

	// Set not default options
	cmd.Flags().StringVarP(&o.Username, "username", "u", "", "")
	cmd.Flags().StringVarP(&o.Password, "password", "p", "", "")
	cmd.Flags().StringVarP(&o.Database, "database", "d", "default", "")
	cmd.Flags().IntVarP(&o.Timeout, "timeout", "", 5, "")
	cmd.Flags().StringVarP(&o.Query, "query", "q", "", "SQL query to execute after auth")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCmdClickhouse())
}
