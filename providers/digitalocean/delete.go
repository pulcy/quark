package digitalocean

import (
	"github.com/pulcy/quark/providers"
)

func (this *doProvider) DeleteCluster(info *providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	droplets, err := this.getInstances(info)
	if err != nil {
		return err
	}
	client := NewDOClient(this.token)
	for _, d := range droplets {
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

func (dp *doProvider) DeleteInstance(info *providers.ClusterInstanceInfo, dnsProvider providers.DnsProvider) error {
	fullName := info.String()
	droplets, err := dp.getInstances(&info.ClusterInfo)
	if err != nil {
		return err
	}
	client := NewDOClient(dp.token)
	for _, d := range droplets {
		if d.Name == fullName {
			// Delete DNS instance records
			instance := dp.clusterInstance(d)
			if err := providers.UnRegisterInstance(dp.Logger, dnsProvider, instance, info.Domain); err != nil {
				return maskAny(err)
			}

			// Delete droplet
			_, err := client.Droplets.Delete(d.ID)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return maskAny(NotFoundError)
}
