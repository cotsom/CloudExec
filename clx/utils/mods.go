package utils

import (
	"fmt"
	"net/netip"
	"os"
)

func ParseTargets(arg string) []netip.Addr {
	var targets []netip.Addr

	ip, err := netip.ParseAddr(arg)
	if err == nil {
		return append(targets, ip)
	}

	prefix, err := netip.ParsePrefix(arg)
	if err != nil {
		fmt.Println("Enter correct host or subnetwork")
		os.Exit(0)
	}

	for addr := prefix.Addr(); prefix.Contains(addr); addr = addr.Next() {
		targets = append(targets, addr)
	}

	if len(targets) < 2 {
		return targets
	}

	return targets[1 : len(targets)-1]
}
