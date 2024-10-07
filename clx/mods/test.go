package main

import "fmt"

type mode string

func (m mode) Run(args []string) {
	fmt.Println("Hello Universe")
}

// экспортируется как символ с именем "Mode"
var Mode mode
