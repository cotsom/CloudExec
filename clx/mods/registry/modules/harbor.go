package modules

import (
	"clx/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Harbor struct {
	Repository []Repository `json:"repository"`
}

type Repository struct {
	ArtifactCount  int    `json:"artifact_count"`
	ProjectID      int    `json:"project_id"`
	ProjectName    string `json:"project_name"`
	ProjectPublic  bool   `json:"project_public"`
	PullCount      int    `json:"pull_count"`
	RepositoryName string `json:"repository_name"`
}

func (m Harbor) RunModule(target string, flags map[string]string, scheme string) {
	port := flags["port"]
	var images Harbor

	client := http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := fmt.Sprintf("%s://%s:%s@%s:%s/", scheme, flags["user"], flags["password"], target, port)
	response, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		return
	}
	respBody, err := ioutil.ReadAll(response.Body)

	if !strings.Contains(string(respBody), "harbor") {
		return
	}

	utils.Colorize(utils.ColorBlue, fmt.Sprintf("%s[*] %s:%s - Harbor\n", utils.ClearLine, target, port))

	url = fmt.Sprintf("%s://%s:%s@%s:%s/api/v2.0/search?q=/", scheme, flags["user"], flags["password"], target, port)

	response, err = utils.HttpRequest(url, http.MethodGet, []byte(""), client)
	if err != nil {
		return
	}

	if response.StatusCode == 401 {
		utils.Colorize(utils.ColorRed, fmt.Sprintf("%s[-] %s:%s - Harbor - %s:%s\n", utils.ClearLine, target, port, flags["user"], flags["password"]))
	}

	respBody, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
	}

	err = json.Unmarshal(respBody, &images)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, image := range images.Repository {
		utils.Colorize(utils.ColorYellow, fmt.Sprintf("[+] %s - %s (Artifacts: %d, Pulls: %d)\n", target, image.RepositoryName, image.ArtifactCount, image.PullCount))
	}
}
