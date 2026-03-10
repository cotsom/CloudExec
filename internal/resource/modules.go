package resource

type ModuleIface interface {
	Run(target string)
	GetName() string
	GetDescription() string
}

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
