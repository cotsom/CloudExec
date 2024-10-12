package utils

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func GetModulesName(path string) ([]string, error) {
	var modules []string
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error:", err)
		return modules, err
	}

	for _, file := range files {
		modules = append(modules, file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))])
	}

	return modules, nil
}

func Contains(s []string, element string) bool {
	for _, a := range s {
		if a == element {
			return true
		}
	}
	return false
}
