package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func GetModulesName(path string) []string {
	var modules []string
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error:", err)
		return modules
	}

	for _, file := range files {
		modules = append(modules, file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))])
	}

	return modules
}

func Contains(s []string, element string) bool {
	for _, a := range s {
		if a == element {
			return true
		}
	}
	return false
}

func GetParam(args []string, moduleSymbol string) (string, error) {
	for i, arg := range args {
		if arg == moduleSymbol {
			if len(args) != i+1 {
				return args[i+1], nil
			}
			err := errors.New("doesn't have param value")
			return "", err
		}
	}
	return "", nil
}
