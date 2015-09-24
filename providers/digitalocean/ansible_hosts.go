package digitalocean

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
)

func (this *doProvider) CreateAnsibleHosts(domain string, sshPort int, developersJson string) error {
	sections := make(map[string]*hostsSection)

	// Load developer offices
	developers, err := loadDevelopOffices(developersJson)
	if err != nil {
		return err
	}
	devSec := &hostsSection{}
	sections["developers"] = devSec

	for k, v := range developers {
		entry := devSec.AddEntry()
		entry.ip = v.Ipv4
		entry.name = k
		entry.region = "developer-office"
		entry.port = sshPort
	}

	// Load droplets
	client := NewDOClient(this.token)
	droplets, err := DropletList(client)
	if err != nil {
		return err
	}
	for _, d := range droplets {
		name := d.Name
		if !strings.Contains(name, domain) {
			continue
		}
		publicV4 := getIpV4(&d, "public")
		privateV4 := getIpV4(&d, "private")
		if publicV4 == "" {
			continue
		}
		groups := []string{"droplets"}
		if d.Region != nil {
			groups = append(groups, d.Region.Slug)
		}
		if strings.Contains(name, "mongo") {
			groups = append(groups, "mongo")
		} else if strings.Contains(name, "ws") {
			groups = append(groups, "webservers")
		} else if strings.Contains(name, "old") {
			groups = append(groups, "oldservers")
		} else if strings.Contains(name, "dev") {
			groups = append(groups, "devservers")
		}
		if strings.Contains(name, "arvika") {
			groups = append(groups, "registry")
		}
		if strings.HasPrefix(name, "ws0.") || strings.HasPrefix(name, "dev0.") || strings.HasPrefix(name, "old.") {
			groups = append(groups, "loadbalancers")
		}
		if strings.HasPrefix(name, "mongo1.") {
			groups = append(groups, "mongo1")
		}

		for _, g := range groups {
			list, ok := sections[g]
			if !ok {
				list = &hostsSection{}
				sections[g] = list
			}
			entry := list.AddEntry()
			entry.ip = publicV4
			entry.region = d.Region.Slug
			entry.name = getSimpleName(name)
			entry.port = sshPort
			if privateV4 != "" {
				entry.privateAddressV4 = privateV4
				entry.privateName = "p" + getSimpleName(name)
			}
		}
	}

	// Print results
	for k, v := range sections {
		fmt.Printf("[%s]\n", k)
		for _, entry := range v.entries {
			fmt.Printf("%s\n", entry.String())
		}
		fmt.Println()
	}
	fmt.Println("[allservers:children]")
	if _, ok := sections["webservers"]; ok {
		fmt.Println("webservers")
	}
	if _, ok := sections["devservers"]; ok {
		fmt.Println("devservers")
	}
	if _, ok := sections["oldservers"]; ok {
		fmt.Println("oldservers")
	}

	return nil
}

func getSimpleName(name string) string {
	return strings.Split(name, ".")[0]
}

func getIpV4(d *godo.Droplet, networkType string) string {
	if d.Networks == nil {
		return ""
	}
	if d.Networks.V4 == nil {
		return ""
	}
	for _, n := range d.Networks.V4 {
		if n.Type == networkType {
			return n.IPAddress
		}
	}
	return ""
}

func loadDevelopOffices(developersJson string) (map[string]*developerEntry, error) {
	bytes, err := ioutil.ReadFile(developersJson)
	if err != nil {
		return nil, err
	}
	data := make(map[string]*developerEntry)
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type developerEntry struct {
	Ipv4 string `json:"ipv4"`
}

type hostsSection struct {
	entries []*hostEntry
}

func (this *hostsSection) AddEntry() *hostEntry {
	if this.entries == nil {
		this.entries = []*hostEntry{}
	}
	entry := &hostEntry{}
	this.entries = append(this.entries, entry)
	return entry
}

type hostEntry struct {
	ip               string
	port             int
	region           string
	name             string
	privateAddressV4 string
	privateName      string
}

// Format the entry as a single line entry
func (this *hostEntry) String() string {
	line := this.name
	line = line + " ansible_ssh_host=" + this.ip
	//var line = this.ip;
	//if (port) {
	//	line = line + ':' + port;
	//}
	line = line + " ansible_ssh_user=admin"
	if this.port != 0 {
		line = line + " ansible_ssh_port=" + strconv.Itoa(this.port)
	}
	line = line + " region=" + this.region
	if this.privateName != "" {
		line = line + " private_name=" + this.privateName
	}
	if this.privateAddressV4 != "" {
		line = line + " private_address_v4=" + this.privateAddressV4
	}
	line = line + " public_address_v4=" + this.ip
	return line
}
