package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

type ModePlugin interface {
	Run(args []string, module string)
}

type Color string

const (
	ColorBlack  Color = "\u001b[30m"
	ColorRed          = "\u001b[31m"
	ColorGreen        = "\u001b[32m"
	ColorYellow       = "\u001b[33m"
	ColorBlue         = "\u001b[34m"
	ColorReset        = "\u001b[0m"
)

func colorize(color Color, message string) {
	fmt.Println(string(color), message, string(ColorReset))
}

func main() {
	M := flag.String("M", "", "choose module")
	flag.Parse()
	fmt.Println(*M)

	if len(os.Args) < 2 {
		fmt.Println("choose mod")
		os.Exit(0)
	}

	mode := os.Args[1]

	pluginDir := "mods/"
	plugins, err := loadPlugins(pluginDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	if plugin, found := plugins[mode]; found {
		fmt.Println(*M)
		if *M != "" {
			plugin.Run(os.Args[2:], *M)
		}
		plugin.Run(os.Args[2:], "")
	} else {
		fmt.Println("Uknown mode: ", mode)
		os.Exit(0)
	}

}

func loadPlugins(pluginDir string) (map[string]ModePlugin, error) {
	plugins := make(map[string]ModePlugin)

	err := filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".so" {
			p, err := plugin.Open(path)
			if err != nil {
				return fmt.Errorf("Can't load plugin %s: %w", path, err)
			}

			symbol, err := p.Lookup("Mode")
			if err != nil {
				return fmt.Errorf("plugin %s doesn't export symbol Mode: %w", path, err)
			}

			mode, ok := symbol.(ModePlugin)
			if !ok {
				return fmt.Errorf("символ Mode в плагине %s не реализует интерфейс ModePlugin", path)
			}

			modeName := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
			plugins[modeName] = mode
		}
		return nil
	})

	return plugins, err
}
