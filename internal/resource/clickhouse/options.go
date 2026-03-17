package click_resources

import "github.com/cotsom/CloudExec/internal/resource"

type ClickhouseOptions struct {
	resource.Options

	// Some default values for clickhouse
	Username string
	Password string
	Database string

	Query   string
	File    string
	Command string
	URL     string

	Timeout int
}
