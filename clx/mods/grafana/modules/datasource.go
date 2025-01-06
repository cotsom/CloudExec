package modules

import (
	"clx/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Datasource struct{}

func (m Datasource) RunModule(targets []string, flags map[string]string) {
	if flags["user"] == "" && flags["password"] == "" {
		grafanaDefaultCreds := [3]string{"admin:admin", "admin:prom-operator", "admin:openbmp"}
		var port string

		if flags["port"] == "" {
			port = "3000"
		} else {
			port = flags["port"]
		}

		client := http.Client{
			Timeout: 1 * time.Second,
		}

		for target := range targets {
			url := fmt.Sprintf("http://%s@%s:%s/api/datasources", creds, target, port)

			response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
			if err != nil {
				return
			}
			respBody, err := ioutil.ReadAll(response.Body)
			defer response.Body.Close()
			if err != nil {
				fmt.Printf("client: could not read response body: %s\n", err)
			}
		}
	}

	fmt.Println("U have called module1 with args", targets)
}
