package modules

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Clone struct{}

func (m Clone) RunModule(target string, flags map[string]string, scheme string) {
	var projects []Project

	port := "80"
	if flags["port"] != "" {
		port = flags["port"]
	}

	body := getProjects(target, flags, scheme, port)
	err := json.Unmarshal(body, &projects)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	for _, project := range projects {
		token := fmt.Sprintf("://oauth2:%s@", flags["token"])
		cloneUrl := strings.Replace(project.Url, "://", token, 1)

		cmd := exec.Command("git", "clone", cloneUrl)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("Output:", string(output))
			return
		}
		fmt.Println(project.Name, string(output))
	}
}
