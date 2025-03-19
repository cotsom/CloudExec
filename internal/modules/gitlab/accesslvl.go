package modules

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type Accesslvl struct {
	Accesslvl int    `json:"access_level"`
	Username  string `json:"name"`
}

type Project struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Username string `json:"name"`
}

func (m Accesslvl) RunModule(target string, flags map[string]string, scheme string) {
	var projects []Project
	var user User
	var access_levels []Accesslvl

	port := "80"
	if flags["port"] != "" {
		port = flags["port"]
	}

	username := getUsername(target, flags, scheme, port)
	err := json.Unmarshal(username, &user)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
	// fmt.Println(user.Username)

	body := getProjects(target, flags, scheme, port)
	err = json.Unmarshal(body, &projects)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	for _, project := range projects {
		fmt.Println("YOUR ACCESS LEVEL FOR PROJECT:", project.Name)

		body = checkPermissions(target, flags, scheme, port, project.Id)

		fmt.Println(string(body))

		err := json.Unmarshal(body, &access_levels)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
		}

		for _, access_level := range access_levels {
			// if access_level.Username == string(username) {
			fmt.Println(access_level.Accesslvl)
			fmt.Println(access_level.Username)
			// }
		}

	}
}

func getUsername(target string, flags map[string]string, scheme string, port string) []byte {
	route := "api/v4/user"

	if flags["timeout"] == "" {
		flags["timeout"] = "10"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	url := fmt.Sprintf("%s://%s:%s/%s", scheme, target, port, route)

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("PRIVATE-TOKEN", flags["token"])
	if err != nil {
		fmt.Println(err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	return body
}

func getProjects(target string, flags map[string]string, scheme string, port string) []byte {
	route := "api/v4/projects?membership=true&per_page=99999"

	if flags["timeout"] == "" {
		flags["timeout"] = "10"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	url := fmt.Sprintf("%s://%s:%s/%s", scheme, target, port, route)

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("PRIVATE-TOKEN", flags["token"])
	if err != nil {
		fmt.Println(err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	return body
}

func checkPermissions(target string, flags map[string]string, scheme string, port string, projectId int) []byte {
	url := fmt.Sprintf("%s://%s:%s/api/v4/projects/%d/members/all", scheme, target, port, projectId)

	if flags["timeout"] == "" {
		flags["timeout"] = "10"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("PRIVATE-TOKEN", flags["token"])
	if err != nil {
		fmt.Println(err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(string(body))
	return body
}
