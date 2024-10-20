package utils

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

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

func CheckPortOpen(host string, port string) {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		fmt.Println("Connecting error:", err)
	}
	if conn != nil {
		defer conn.Close()
		fmt.Println("Opened", net.JoinHostPort(host, port))
	}
}

func ValidIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

	return re.MatchString(ipAddress)
}

type Color string

const (
	ColorBlack  Color = "\u001b[30m"
	ColorRed    Color = "\u001b[31m"
	ColorGreen  Color = "\u001b[32m"
	ColorYellow Color = "\u001b[33m"
	ColorBlue   Color = "\u001b[34m"
	ColorReset  Color = "\u001b[0m"
)

func Colorize(color Color, message string) {
	fmt.Println(string(color), message, string(ColorReset))
}
