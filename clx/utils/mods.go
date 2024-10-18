package utils

import (
	"fmt"
	"log"
	"net"
)

func ParseTargets(arg string) []string {
	var targets []string

	ipv4Addr, ipv4Net, err := net.ParseCIDR("192.0.2.1/24")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ipv4Addr)
	fmt.Println(ipv4Net)

	if ValidIP4(arg) {
		targets = append(targets, arg)
		return targets
	}

	return nil
}
