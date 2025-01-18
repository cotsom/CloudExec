package utils

import (
	"bufio"
	"fmt"
	"log"
	"net/netip"
	"os"
	"regexp"
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,}$`)

func ParseTargets(arg string) []string {
	var targets []string

	//parse ip
	ip, err := netip.ParseAddr(arg)
	if err == nil {
		return append(targets, ip.String())
	}

	//parse domain
	if domainRegex.MatchString(arg) {
		return append(targets, arg)
	}

	//parse network
	prefix, err := netip.ParsePrefix(arg)
	if err != nil {
		fmt.Println("Enter correct host or subnetwork")
		os.Exit(0)
	}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {
		targets = append(targets, addr.String())
	}

	if len(targets) < 2 {
		return targets
	}

	//Delete .0 and .255 targets
	for i := 0; i < 2; i++ {
		first_target := []rune(targets[0])
		if (string(first_target[len(first_target)-2:]) == ".0") || (string(first_target[len(first_target)-3:]) == "255") {
			targets[0] = targets[len(targets)-1] // Copy last element to index i.
			targets = targets[:len(targets)-1]   // Truncate slice.
		}
	}

	return targets
}

func ParseTargetsFromList(inputFile string) []string {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}
