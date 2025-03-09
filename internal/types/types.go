package types

type Module interface {
	RunModule(target string, flags map[string]string)
}
