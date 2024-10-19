package main

import (
	"clx/utils"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

type ModePlugin interface {
	Run(args []string)
}

func main() {
	if len(os.Args) < 2 {
		utils.Colorize(utils.ColorRed, "choose mode")
		os.Exit(0)
	}

	mode := os.Args[1]

	currentPath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		return
	}
	workingDir := filepath.Dir(currentPath)

	pluginDir := fmt.Sprintf("%s/mods/%s", workingDir, mode)
	plugins, err := loadPlugins(pluginDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	if plugin, found := plugins[mode]; found {
		plugin.Run(os.Args[2:])
	} else {
		fmt.Println("Uknown mode: ", mode)
		os.Exit(0)
	}

}

func loadPlugins(pluginDir string) (map[string]ModePlugin, error) {
	plugins := make(map[string]ModePlugin)

	err := filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			mode := strings.Split(pluginDir, "/")
			utils.Colorize(utils.ColorRed, fmt.Sprint("Uknown mode: ", mode[len(mode)-1]))
			os.Exit(0)
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
