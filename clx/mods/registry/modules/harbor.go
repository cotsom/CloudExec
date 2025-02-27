package modules

import (
	"clx/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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

type Artifact struct {
	AdditionLinks struct {
		BuildHistory struct {
			Href string `json:"href"`
		} `json:"build_history"`
	} `json:"addition_links"`
}

func (m Harbor) RunModule(target string, flags map[string]string, scheme string) {
	port := flags["port"]
	var images Harbor
	var artifacts []Artifact

	if flags["timeout"] == "" {
		flags["timeout"] = "1"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
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
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("[+] %s - %s (Artifacts: %d, Pulls: %d)\n", target, image.RepositoryName, image.ArtifactCount, image.PullCount))
		repoNameSplit := strings.SplitN(image.RepositoryName, "/", 2)

		url := fmt.Sprintf("%s://%s:%s@%s:%s/api/v2.0/projects/%s/repositories/%s/artifacts?with_tag=false&with_scan_overview=true&with_label=true&with_accessory=false&page_size=15&page=1", scheme, flags["user"], flags["password"], target, port, repoNameSplit[0], strings.ReplaceAll(repoNameSplit[1], "/", "%252F"))
		response, err = utils.HttpRequest(url, http.MethodGet, []byte(""), client)
		if err != nil {
			return
		}
		respBody, err = ioutil.ReadAll(response.Body)
		defer response.Body.Close()
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}

		err := json.Unmarshal(respBody, &artifacts)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		for _, artifact := range artifacts {
			url := fmt.Sprintf("%s://%s:%s@%s:%s/%s", scheme, flags["user"], flags["password"], target, port, artifact.AdditionLinks.BuildHistory.Href)
			response, err = utils.HttpRequest(url, http.MethodGet, []byte(""), client)
			if err != nil {
				return
			}
			respBody, err = ioutil.ReadAll(response.Body)
			defer response.Body.Close()
			if err != nil {
				fmt.Printf("client: could not read response body: %s\n", err)
			}

			var data interface{}
			err = json.Unmarshal(respBody, &data)
			if err != nil {
				fmt.Println("Ошибка при декодировании JSON:", err)
				// fmt.Println(string(respBody))
				// return
			}

			prettyJSON, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				fmt.Println("Ошибка при форматировании JSON:", err)
				return
			}

			fmt.Println(string(prettyJSON))
			utils.Colorize(utils.ColorYellow, fmt.Sprintf("%s\n", string(prettyJSON)))
		}
	}
}
