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

	//Delete .0 and .255 targets
	for i := 0; i < 2; i++ {
		first_target := []rune(targets[0].String())
		if (string(first_target[len(first_target)-1]) == "0") || (string(first_target[len(first_target)-3:]) == "255") {
			targets[0] = targets[len(targets)-1] // Copy last element to index i.
			targets = targets[:len(targets)-1]   // Truncate slice.
		}
	}

	return targets
}
