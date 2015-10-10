package vultr

import (
	"arvika.pulcy.com/pulcy/droplets/providers"
)

// Remove all instances of a cluster
func (vp *vultrProvider) DeleteCluster(info *providers.ClusterInfo, dnsProvider providers.DnsProvider) error {
	return maskAny(NotImplementedError)
}
