package modules

import (
	"fmt"

	"github.com/cotsom/CloudExec/internal/resource"
	clickResources "github.com/cotsom/CloudExec/internal/resource/clickhouse"
)

type ClickhouseBruteModule struct {
	resource.Module

	Opts clickResources.ClickhouseOptions
}

func (m *ClickhouseBruteModule) Run(target string) {
	fmt.Println(target)
}

func NewClickhouseBruteModule(opts clickResources.ClickhouseOptions) *ClickhouseBruteModule {
	module := &ClickhouseBruteModule{
		Module: resource.Module{
			Name:        "default",
			Description: "Checks default credentials",
		},

		Opts: opts,
	}

	return module
}
