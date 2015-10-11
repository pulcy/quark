package vultr

import (
	"errors"
	"strings"

	"arvika.pulcy.com/pulcy/droplets/providers"
)

// Remove all instances of a cluster
func (vp *vultrProvider) DeleteCluster(info *providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	servers, err := vp.getInstances(info)
	if err != nil {
		return err
	}
	for _, s := range servers {
		if strings.Contains(s.Name, "arvika") {
			return errors.New("Not allowed to delete arvika")
		}

		// Delete DNS instance records
		instance := vp.clusterInstance(s)
		if err := providers.UnRegisterInstance(vp.Logger, dnsProvider, instance, info.Domain); err != nil {
			return maskAny(err)
		}

		// Delete droplet
		err := vp.client.DeleteServer(s.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
