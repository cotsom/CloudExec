package cmd

import (
	"database/sql"
	"fmt"
	"strings"

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
		Command: resource.Command{
			Logger: resource.NewLogger(),
		},
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

	var query string
	switch {
	case c.Opts.Query != "":
		query = c.Opts.Query
	case c.Opts.Command != "":
		query = fmt.Sprintf("SELECT * FROM executable('%s', 'LineAsString', 'result String')", c.Opts.Command)
	case c.Opts.File != "":
		query = fmt.Sprintf("SELECT * FROM file('%s', 'LineAsString', 'result String')", c.Opts.File)
	case c.Opts.URL != "":
		if !strings.HasPrefix(c.Opts.URL, "http://") && !strings.HasPrefix(c.Opts.URL, "https://") {
			c.Opts.URL = "http://" + c.Opts.URL
		}

		headers := ""
		if len(c.Opts.Headers) > 0 {
			for i, header := range c.Opts.Headers {
				delim := strings.Index(header, ":")
				headerKey := strings.TrimSpace(header[:delim])
				headerValue := strings.TrimSpace(header[delim+1:])

				headers = fmt.Sprintf("%s'%s'='%s'", headers, headerKey, headerValue)
				if i != len(c.Opts.Headers)-1 {
					headers = fmt.Sprintf("%s, ", headers)
				}
			}
			headers = fmt.Sprintf(", headers(%s)", headers)
		}
		query = fmt.Sprintf("SELECT * FROM url('%s', 'LineAsString', 'result String'%s)", c.Opts.URL, headers)
	default:
		return nil
	}

	rows, err := conn.ExecuteQuery(query)
	if err != nil {
		c.Logger.Fatal(err.Error())
		return nil
	}
	defer rows.Close()
	c.Logger.Raw(conn.PrintableRows(rows))

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
	cmd.Flags().IntVarP(&c.Opts.Port, "port", "P", 9000, "Clickhouse port")

	// Set not default options
	cmd.Flags().StringVarP(&c.Opts.Username, "username", "u", "", "")
	cmd.Flags().StringVarP(&c.Opts.Password, "password", "p", "", "")
	cmd.Flags().StringVarP(&c.Opts.Database, "database", "d", "default", "Database to connect in clickhouse")
	cmd.Flags().IntVarP(&c.Opts.Timeout, "timeout", "", 5, "Clickhouse connection timeout")

	cmd.Flags().StringVarP(&c.Opts.Query, "query", "q", "", "SQL query to execute after auth")
	cmd.Flags().StringVarP(&c.Opts.URL, "ssrf-url", "U", "", "URL for GET SSRF")
	cmd.Flags().StringArrayVarP(&c.Opts.Headers, "ssrf-header", "H", []string{}, "Headers for GET SSRF")
	cmd.Flags().StringVarP(&c.Opts.File, "read-file", "F", "", "File to read in <user_files_path> configuration folder")
	cmd.Flags().StringVarP(&c.Opts.Command, "command", "x", "", "Command to execute from <user_scripts_path> configuration folder")

	// Modules
	c.RegisterModule(modules.NewClickhouseBruteModule(c.Opts, c.Logger))

	return cmd
}

func init() {
	rootCmd.AddCommand(NewCmdClickhouse())
}
