package modules

import (
	"database/sql"
	"fmt"

	"github.com/cotsom/CloudExec/internal/resource"
	clickResources "github.com/cotsom/CloudExec/internal/resource/clickhouse"
	"github.com/cotsom/CloudExec/internal/utils/sqlquery"
)

var defaultCreds map[string]string = map[string]string{
	"default": "default",
}

type ClickhouseBruteModule struct {
	resource.Module

	Opts   clickResources.ClickhouseOptions
	Logger resource.Logger
}

func (m *ClickhouseBruteModule) checkConnection(target, username, password string) {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=%ds&read_timeout=%ds",
		username,
		password,
		target,
		m.Opts.Port,
		m.Opts.Database,

		m.Opts.Timeout,
		m.Opts.Timeout,
	)

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		m.Logger.Fatal(err.Error())
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
			m.Logger.Error(fmt.Sprintf("Cickhouse: %s - %s:%s", target, username, password))
		case sqlquery.DatabaseFailure:
			m.Logger.Error(fmt.Sprintf("Cickhouse: %s - %s:%s\tDatabase doesn't exist: %s", target, username, password, m.Opts.Database))
		default:
			m.Logger.Fatal(err.Error())
		}
		return
	}

	m.Logger.Found(fmt.Sprintf("Cickhouse: %s - %s:%s", target, username, password))
}

func (m *ClickhouseBruteModule) Run(target string) {
	for username, password := range defaultCreds {
		m.checkConnection(target, username, password)
	}

}

func NewClickhouseBruteModule(opts clickResources.ClickhouseOptions, logger resource.Logger) *ClickhouseBruteModule {
	module := &ClickhouseBruteModule{
		Module: resource.Module{
			Name:        "default-creds",
			Description: "Checks default credentials",
		},
		Logger: logger,
		Opts:   opts,
	}

	module.Module.ModuleIface = module

	return module
}
