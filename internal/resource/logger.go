package resource

import (
	"fmt"

	"github.com/cotsom/CloudExec/internal/utils"
)

// TODO: Thing about --no-color
type Logger struct{}

func (l Logger) Info(text string) {
	utils.Colorize(
		utils.ColorBlue,
		fmt.Sprintf("[*] %s", text),
	)
}

func (l Logger) Found(text string) {
	utils.Colorize(
		utils.ColorGreen,
		fmt.Sprintf("[+] %s", text),
	)
}

func (l Logger) Error(text string) {
	utils.Colorize(
		utils.ColorRed,
		fmt.Sprintf("[-] %s", text),
	)
}

func (l Logger) Fatal(text string) {
	utils.Colorize(
		utils.ColorYellow,
		fmt.Sprintf("[!!!] %s", text),
	)
}

func (l Logger) Raw(text string) {
	fmt.Println(text)
}

func (l Logger) List(text string) {
	fmt.Printf("-> %s\n", text)
}
