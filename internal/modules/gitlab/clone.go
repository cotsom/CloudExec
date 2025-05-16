package modules

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cotsom/CloudExec/internal/utils"
)

type Clone struct{}

func (m Clone) RunModule(target string, flags map[string]string, scheme string) {
	var projects []Project

	port := "80"
	if flags["port"] != "" {
		port = flags["port"]
	}

	if flags["public"] == "true" {
		fmt.Println(flags["public"])
		body, err := getPublicProjects(target, flags, scheme, port)
		if err != nil {
			fmt.Println("Error getting projects:", err)
		}

		err = json.Unmarshal(body, &projects)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
		}

		for _, project := range projects {
			cmd := exec.Command("git", "clone", project.Url)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println("Error:", err)
				fmt.Println("Output:", string(output))
				return
			}
			fmt.Println(project.Name, string(output))
		}
	}

	body, err := getProjects(target, flags, scheme, port)
	if err != nil {
		fmt.Println("Error getting projects:", err)
	}

	err = json.Unmarshal(body, &projects)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	for _, project := range projects {
		token := fmt.Sprintf("://oauth2:%s@", flags["token"])
		cloneUrl := strings.Replace(project.Url, "://", token, 1)

		cmd := exec.Command("git", "clone", cloneUrl)
		output, err := cmd.CombinedOutput()
		if err != nil {
			utils.Colorize(utils.ColorRed, fmt.Sprintf("Error: %s", err))
			fmt.Println("Output:", string(output))
		}
		fmt.Println(project.Name, string(output))
	}
}
