package providers

import (
	"github.com/op/go-logging"
)

// UpdateClusterMembers updates /etc/yard-cluster-members on all instances of the cluster
func UpdateClusterMembers(log *logging.Logger, info ClusterInfo, provider CloudProvider) error {
	// See if there are already instances for the given cluster
	instances, err := provider.GetInstances(&info)
	if err != nil {
		return maskAny(err)
	}

	// Update existing members
	clusterMembers := []ClusterMember{}
	for _, i := range instances {
		machineID, err := i.GetMachineID(log)
		if err != nil {
			return maskAny(err)
		}
		clusterMembers = append(clusterMembers, ClusterMember{MachineID: machineID, PrivateIP: i.PrivateIpv4})
	}
	for _, i := range instances {
		if err := i.UpdateClusterMembers(log, clusterMembers); err != nil {
			return maskAny(err)
		}
	}
	return nil
}
