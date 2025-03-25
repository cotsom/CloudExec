package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/rand"
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
	ColorBlack  Color  = "\u001b[30m"
	ColorRed    Color  = "\u001b[31m"
	ColorGreen  Color  = "\u001b[32m"
	ColorYellow Color  = "\u001b[33m"
	ColorBlue   Color  = "\u001b[34m"
	ColorReset  Color  = "\u001b[0m"
	ClearLine   string = "\033[2K\n"
)

func Colorize(color Color, message string) {
	fmt.Println(string(color), message, string(ColorReset))
}

func HttpRequest(targetUrl string, method string, data []byte, client http.Client) (*http.Response, error) {
	request, err := http.NewRequest(method, targetUrl, bytes.NewBuffer(data))

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response, err := client.Do(request)
	if err != nil {
		return response, err
	}

	return response, nil
}

func ProgressBar(allItems int, currentItem int, progress *int) {
	percent := (currentItem * 100) / allItems

	// bar := percent - *progress
	// *progress = percent

	fmt.Printf("\r[%s] %d", strings.Repeat("-", percent), percent)
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
