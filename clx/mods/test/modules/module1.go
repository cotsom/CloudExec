package modules

import "fmt"

type Module1 struct{}

func (m Module1) RunModule(arg string) {
	fmt.Println("U have called module1 with args", arg)
}
