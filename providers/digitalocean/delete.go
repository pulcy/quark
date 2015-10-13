package digitalocean

import (
	"errors"
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
		instance := this.clusterInstance(d)
		if err := providers.UnRegisterInstance(this.Logger, dnsProvider, instance, info.Domain); err != nil {
			return maskAny(err)
		}

		// Delete droplet
		_, err := client.Droplets.Delete(d.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
