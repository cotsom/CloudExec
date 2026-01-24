package modules

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	utils "github.com/cotsom/CloudExec/internal/utils"
)

type Accesslvl struct {
	Accesslvl int    `json:"access_level"`
	Username  string `json:"name"`
}

type SharedWithGroup struct {
	GroupAccessLevel int `json:"group_access_level"`
}

type Project struct {
	Id                int               `json:"id"`
	Name              string            `json:"name"`
	Permissions       Permissions       `json:"permissions"`
	Url               string            `json:"http_url_to_repo"`
	SharedWithGroups  []SharedWithGroup `json:"shared_with_groups"`
	PathWithNamespace string            `json:"path_with_namespace"`
}

type Permissions struct {
	ProjectAccess *GroupAccess `json:"project_access"`
	GroupAccess   *GroupAccess `json:"group_access"`
}

type GroupAccess struct {
	AccessLevel int `json:"access_level"`
}

type User struct {
	Id       int    `json:"id"`
	Username string `json:"name"`
}

func (m Accesslvl) RunModule(target string, flags map[string]string, scheme string) {
	var projects []Project
	var user User
	var access_level Accesslvl

	port := "80"
	if flags["port"] != "" {
		port = flags["port"]
	}

	username, err := getUsername(target, flags, scheme, port)
	if err != nil {
		fmt.Println("Error getting user:", err)
	}

	err = json.Unmarshal(username, &user)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	body, err := getProjects(target, flags, scheme, port)
	if err != nil {
		fmt.Println("Error getting projects:", err)
	}

	err = json.Unmarshal(body, &projects)
	if err != nil {
		fmt.Println("Can't read projects", string(body))
	}

	for _, project := range projects {
		utils.Colorize(
			utils.ColorYellow,
			fmt.Sprintf("===========%s==========", project.Name),
		)

		printGroupAccess(project)

		body, err := checkPermissions(target, flags, scheme, port, project.Id, user.Id)
		if err != nil {
			fmt.Printf("Error getting permissions for project %d: %v\n", project.Id, err)
			continue
		}

		if err = json.Unmarshal(body, &access_level); err != nil {
			fmt.Printf("Error unmarshalling JSON for project %d: %v\n", project.Id, err)
			continue
		}

		utils.Colorize(
			utils.ColorYellow,
			fmt.Sprintf(
				"User %s ACCESS LEVEL FOR PROJECT: %s (ID: %d)",
				access_level.Username,
				project.Name,
				project.Id,
			),
		)

		utils.Colorize(
			utils.ColorGreen,
			fmt.Sprintf("%d", access_level.Accesslvl),
		)
	}
}

func makeRequest(url, token string, timeout int) ([]byte, error) {
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("PRIVATE-TOKEN", token)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

func getUsername(target string, flags map[string]string, scheme, port string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/api/v4/user", scheme, target, port)
	return makeRequest(url, flags["token"], utils.GetTimeout(flags))
}

func getProjects(target string, flags map[string]string, scheme, port string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/api/v4/projects?membership=true&per_page=50000", scheme, target, port)
	return makeRequest(url, flags["token"], utils.GetTimeout(flags))
}

func getPublicProjects(target string, flags map[string]string, scheme, port string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/api/v4/projects?per_page=50000", scheme, target, port)
	return makeRequest(url, flags["token"], utils.GetTimeout(flags))
}

func checkPermissions(target string, flags map[string]string, scheme, port string, projectId int, userId int) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/api/v4/projects/%d/members/all/%d", scheme, target, port, projectId, userId)
	return makeRequest(url, flags["token"], utils.GetTimeout(flags))
}

func getProjectById(target string, flags map[string]string, scheme, port string, projectId int) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/api/v4/projects/%d", scheme, target, port, projectId)
	return makeRequest(url, flags["token"], utils.GetTimeout(flags))
}

func printGroupAccess(project Project) {
	var levels []int

	if ga := project.Permissions.GroupAccess; ga != nil {
		levels = append(levels, ga.AccessLevel)
	} else if groups := project.SharedWithGroups; groups != nil {
		for _, g := range groups {
			levels = append(levels, g.GroupAccessLevel)
		}
	}

	if len(levels) == 0 {
		return
	}

	for _, level := range levels {
		utils.Colorize(utils.ColorYellow,
			fmt.Sprintf("Group Access Level: %d\n", level))
	}
}
