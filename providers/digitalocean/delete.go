package digitalocean

import (
	"errors"
	"fmt"
	"strings"

	"arvika.pulcy.com/pulcy/droplets/providers"
)

func (this *doProvider) DeleteCluster(info *providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	droplets, err := this.getInstances(info)
	if err != nil {
		return err
	}
	client := NewDOClient(this.token)
	for _, d := range droplets {
		if strings.Contains(d.Name, "arvika") {
			return errors.New("Not allowed to delete arvika")
		}

		// Delete DNS instance records
		if err := dnsProvider.DeleteDnsRecord(info.Domain, "A", d.Name, ""); err != nil {
			return err
		}
		if err := dnsProvider.DeleteDnsRecord(info.Domain, "AAAA", d.Name, ""); err != nil {
			return err
		}

		// Delete droplet
		_, err := client.Droplets.Delete(d.ID)
		if err != nil {
			return err
		}
	}

	// Delete DNS cluster records
	clusterName := fmt.Sprintf("%s.%s", info.Name, info.Domain)
	if err := dnsProvider.DeleteDnsRecord(info.Domain, "A", clusterName, ""); err != nil {
		return err
	}
	if err := dnsProvider.DeleteDnsRecord(info.Domain, "AAAA", clusterName, ""); err != nil {
		return err
	}

	return nil
}
