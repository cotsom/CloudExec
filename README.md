## How to create your own module

1. Create your module's directory (clx/mods/mymodule)
2. Create file mymodule.go
3. Implement the interface using the minimal module template

```Go
package main

import (
	modules "clx/mods/test/modules"
)

// mode type for plugin's symbol
type mode string

type Module interface {
	RunModule(arg string)
}

var registeredModules = map[string]Module{
	"module1": modules.Module1{},
	// Add another modules here
}

func (m mode) Run(args []string) {
  //pass
}

// exporting symbol "Mode"
var Mode mode

```

USAGE: clx mymodule 192.168.1.5
