package vultr

import (
	"fmt"
	"strings"

	"github.com/JamesClonk/vultr/lib"

	"arvika.pulcy.com/pulcy/droplets/providers"
)

// Get names of instances of a cluster
func (vp *vultrProvider) GetInstances(info *providers.ClusterInfo) ([]providers.ClusterInstance, error) {
	servers, err := vp.getInstances(info)
	if err != nil {
		return nil, maskAny(err)
	}
	list := []providers.ClusterInstance{}
	for _, s := range servers {
		info := providers.ClusterInstance{
			Name:        s.Name,
			PrivateIpv4: s.InternalIP,
			PublicIpv4:  s.MainIP,
		}
		list = append(list, info)

	}
	return list, nil
}

func (vp *vultrProvider) getInstances(info *providers.ClusterInfo) ([]lib.Server, error) {
	servers, err := vp.client.GetServers()
	if err != nil {
		return nil, maskAny(err)
	}

	postfix := fmt.Sprintf(".%s.%s", info.Name, info.Domain)
	result := []lib.Server{}
	for _, s := range servers {
		if strings.HasSuffix(s.Name, postfix) {
			result = append(result, s)
		}
	}

	return result, nil
}

// clusterInstance creates a ClusterInstance record for the given server
func (dp *vultrProvider) clusterInstance(s lib.Server) providers.ClusterInstance {
	info := providers.ClusterInstance{
		Name:        s.Name,
		PrivateIpv4: s.InternalIP,
		PublicIpv4:  s.MainIP,
		PublicIpv6:  s.MainIPV6,
	}
	return info
}
