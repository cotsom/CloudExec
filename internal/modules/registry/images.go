package modules

import (
	"clx/internal/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Images struct {
	Repositories []string `json:"repositories"`
}

func (m Images) RunModule(target string, flags map[string]string, scheme string) {
	port := flags["port"]
	var images Images

	client := http.Client{
		Timeout: 1 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := fmt.Sprintf("%s://%s:%s@%s:%s/v2/_catalog", scheme, flags["user"], flags["password"], target, port)
	response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		return
	}

	if response.StatusCode == 401 {
		return
	}

	respBody, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
	}

	err = json.Unmarshal(respBody, &images)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, image := range images.Repositories {
		utils.Colorize(utils.ColorYellow, fmt.Sprintf("[+] %s - %s", target, image))
	}

}
