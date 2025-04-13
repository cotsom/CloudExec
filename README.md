# CloudExec - Cloud Execution Tool
This utility is designed to scan, detect, exploit vulnerabilities and services misconfigurations in cloud and dev linux infrastructures. The selected modes were services commonly found in such environments, as well as the search and testing processes I wanted to automate :)

Here are mods for scan some services
- Grafana
- Postgres
- Gitlab
- Zookeeper
- Kafka
- Registry
- Kube


# Usage: 
```shell
#Common service discovery
clx ModeName <ip/network/hostname>

#Take hosts from list
clx ModeName -i hosts.txt

#Using module on all found hosts
clx ModeName <ip/network/hostname> -M moduleName
```

### Legend
Blue highlighting - Target found `[*] 192.168.1.1 - Service`

Green highlighting - Target found and access granted  `[+] 192.168.1.1 - Service`

Yellow highlighted `Pwned` - You can execute code (RCE)

# Grafana
This mode  is designed to discover & exploit Grafana. It will scan and highlight all found hosts with grafana service

**Modules**:
* **datasources** - Displays a list of all available sources for the specified account. By querying the data sources, you can retrieve the data stored in them (require `-u` and `-p` flags for authenticate)
* **defcreds** - Try to authenticate with popular creds


# Postgres
This mode  is designed to discover & exploit Postgres. It will scan and highlight all found hosts. If the creds are correct, then it will automatically check if the user is a superuser and if so, a yellow Pwned label will appear, indicating RCE capability

Code execution: `clx postgres <ip> -x id`


# Install
Go 1.23+

`$ go install github.com/cotsom/CloudExec@latest`

# Writing module
You can add your own module using e.g. `cobra-cli add <commandName>` command.  This will create a new file in the `cmd/` directory containing the code template for the new command.

#### Template for new mode:
test mode

test module

*In the Run function you can implement your own logic for receiving and parsing flags, targets and scans*

