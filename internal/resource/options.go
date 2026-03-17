package resource

// Structure for cobra flags parsing
// Can be embeded by children
type Options struct {
	Inputlist string
	Port      int

	ListModules bool
	Module      string

	Threads int
}
