package digitalocean

import (
	"fmt"
	"strings"

	"arvika.pulcy.com/pulcy/droplets/providers"
	"github.com/digitalocean/godo"
)

func (this *doProvider) GetInstances(info *providers.ClusterInfo) ([]providers.ClusterInstance, error) {
	droplets, err := this.getInstances(info)
	if err != nil {
		return nil, err
	}
	result := []providers.ClusterInstance{}
	for _, d := range droplets {
		info := providers.ClusterInstance{
			Name:        d.Name,
			PrivateIpv4: getIpv4(d, "private"),
			PublicIpv4:  getIpv4(d, "public"),
		}
		result = append(result, info)
	}
	return result, nil
}

func (this *doProvider) getInstances(info *providers.ClusterInfo) ([]godo.Droplet, error) {
	client := NewDOClient(this.token)
	droplets, err := DropletList(client)
	if err != nil {
		return nil, err
	}

	postfix := fmt.Sprintf(".%s.%s", info.Name, info.Domain)
	result := []godo.Droplet{}
	for _, d := range droplets {
		if strings.HasSuffix(d.Name, postfix) {
			result = append(result, d)
		}
	}

	return result, nil
}

func getIpv4(d godo.Droplet, nType string) string {
	if d.Networks == nil {
		return ""
	}
	for _, n := range d.Networks.V4 {
		if n.Type == nType {
			return n.IPAddress
		}
	}
	return ""
}

func getIpv6(d godo.Droplet, nType string) string {
	if d.Networks == nil {
		return ""
	}
	for _, n := range d.Networks.V6 {
		if n.Type == nType {
			return n.IPAddress
		}
	}
	return ""
}
