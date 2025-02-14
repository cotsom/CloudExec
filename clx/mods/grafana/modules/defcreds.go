package modules

import (
	"clx/utils"
	"fmt"
	"net/http"
	"time"
)

type Defcreds struct{}

func (m Defcreds) RunModule(target string, flags map[string]string) {
	grafanaDefaultCreds := [3]string{"admin:admin", "admin:prom-operator", "admin:openbmp"}
	port := "3000"

	if flags["port"] != "" {
		port = flags["port"]
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	for _, creds := range grafanaDefaultCreds {
		url := fmt.Sprintf("http://%s@%s:%s/api/org", creds, target, port)
		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()

		if response.StatusCode == 200 {
			utils.Colorize(utils.ColorGreen, fmt.Sprintf("%s[+] %s - Grafana (%s)", utils.ClearLine, target, creds))
		}
	}
}
