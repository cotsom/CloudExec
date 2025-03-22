package modules

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	utils "github.com/cotsom/CloudExec/internal/utils"
)

type Loginbypass struct{}

func (m Loginbypass) RunModule(target string, flags map[string]string, scheme string) {
	bypassRoutes := [3]string{"explore", "api/v4/projects?visibility=public", "search?search="}
	port := "80"

	if flags["port"] != "" {
		port = flags["port"]
	}

	if flags["timeout"] == "" {
		flags["timeout"] = "10"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	for _, route := range bypassRoutes {
		url := fmt.Sprintf("%s://%s:%s/%s", scheme, target, port, route)
		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()

		if response.StatusCode == 200 {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s - Login bypassed (%s)", utils.ClearLine, target, route))
		}
	}
}
