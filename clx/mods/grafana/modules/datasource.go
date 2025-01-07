package modules

import (
	"clx/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Datasource struct {
	ID          int            `json:"id"`
	UID         string         `json:"uid"`
	OrgID       int            `json:"orgId"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	TypeName    string         `json:"typeName"`
	TypeLogoURL string         `json:"typeLogoUrl"`
	Access      string         `json:"access"`
	URL         string         `json:"url"`
	User        string         `json:"user"`
	Database    string         `json:"database"`
	BasicAuth   bool           `json:"basicAuth"`
	IsDefault   bool           `json:"isDefault"`
	JsonData    map[string]any `json:"jsonData"`
	ReadOnly    bool           `json:"readOnly"`
}

func (m Datasource) RunModule(targets []string, flags map[string]string) {
	if flags["user"] == "" && flags["password"] == "" {
		return
	}

	var datasources []Datasource
	var port string

	if flags["port"] == "" {
		port = "3000"
	} else {
		port = flags["port"]
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	for _, target := range targets {
		url := fmt.Sprintf("http://%s:%s@%s:%s/api/datasources", flags["user"], flags["password"], target, port)

		response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			return
		}
		respBody, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		err = json.Unmarshal(respBody, &datasources)
		if err != nil {
			fmt.Println("Ошибка разбора JSON:", err)
			return
		}

		for _, datasource := range datasources {
			utils.Colorize(utils.ColorYellow, fmt.Sprintf("[*] %s", datasource.Name))
		}

	}
}
