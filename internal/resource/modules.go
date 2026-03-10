package resource

// Run method will be implementeted by children
type ModuleIface interface {
	Run(target string)
	GetName() string
	GetDescription() string
}

// Module parent struct for -M and -L flags
type Module struct {
	ModuleIface

	Name        string
	Description string

	Opts Options
}

func (m *Module) GetName() string {
	return m.Name
}

func (m *Module) GetDescription() string {
	return m.Description
}
