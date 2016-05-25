// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scaleway

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/juju/errgo"
	"github.com/op/go-logging"
	"github.com/scaleway/scaleway-cli/pkg/api"

	"github.com/pulcy/quark/providers"
	"github.com/pulcy/quark/templates"
	"github.com/pulcy/quark/util"
)

const (
	fileMode            = os.FileMode(0775)
	cloudConfigTemplate = "templates/cloud-config.tmpl"
	bootstrapTemplate   = "templates/scaleway-bootstrap.tmpl"
	volumeType          = "l_ssd"
	volumeSize          = uint64(50 * 1000 * 1000 * 1000)

	clusterIDTagIndex = 0 // Index in ScalewayServer.Tags of the cluster-ID
	clusterIPTagIndex = 1 // Index in ScalewayServer.Tags of the cluster IP address (tinc address)
)

// CreateInstance creates one new machine instance.
func (vp *scalewayProvider) CreateInstance(log *logging.Logger, options providers.CreateInstanceOptions, dnsProvider providers.DnsProvider) (providers.ClusterInstance, error) {
	// Fetch existing instances
	existingInstances, err := vp.GetInstances(options.ClusterInfo)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	// Create server
	instance, err := vp.createInstance(log, options, dnsProvider, existingInstances)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	// Update tinc network config
	instanceList, err := vp.GetInstances(options.ClusterInfo)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}
	newInstances := providers.ClusterInstanceList{instance}
	if instanceList.ReconfigureTincCluster(vp.Logger, newInstances); err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	return instance, nil
}

// createInstance creates a new instances, runs the bootstrap script and registers the instance
// in DNS.
func (vp *scalewayProvider) createInstance(log *logging.Logger, options providers.CreateInstanceOptions, dnsProvider providers.DnsProvider, existingInstances providers.ClusterInstanceList) (providers.ClusterInstance, error) {
	// Create a new machine ID
	machineID, err := util.GenUUID()
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}
	log.Debugf("created machined-id: %s", machineID)

	// Create server
	instance, err := vp.createAndStartServer(options)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	// Update `cluster-members` file on existing instances.
	// This ensures that the firewall of the existing instances allows our new instance
	if len(existingInstances) > 0 {
		rebootAfter := false
		clusterMembers, err := existingInstances.AsClusterMemberList(log, nil)
		if err != nil {
			return providers.ClusterInstance{}, maskAny(err)
		}
		newMember := providers.ClusterMember{
			ClusterID:     options.ClusterInfo.ID,
			MachineID:     machineID,
			ClusterIP:     instance.ClusterIP,
			PrivateHostIP: instance.PrivateIP,
			EtcdProxy:     options.EtcdProxy,
		}
		clusterMembers = append(clusterMembers, newMember)
		if err := existingInstances.UpdateClusterMembers(log, clusterMembers, rebootAfter, vp); err != nil {
			log.Warningf("Failed to update cluster members: %#v", err)
		}
	}

	// Bootstrap server
	if err := vp.bootstrapServer(instance, options, machineID); err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	// Wait for the server to be active
	server, err := vp.waitUntilServerActive(instance.ID, false)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	if options.RoleLoadBalancer {
		privateIpv4 := server.PrivateIP
		publicIpv4 := server.PublicAddress.IP
		publicIpv6 := ""
		if server.IPV6 != nil {
			publicIpv6 = server.IPV6.Address
		}
		if err := providers.RegisterInstance(vp.Logger, dnsProvider, options, server.Name, options.RegisterInstance, options.RoleLoadBalancer, options.RoleLoadBalancer, publicIpv4, publicIpv6, privateIpv4); err != nil {
			return providers.ClusterInstance{}, maskAny(err)
		}
	}

	vp.Logger.Infof("Server '%s' is ready", server.Name)

	return vp.clusterInstance(server, false), nil
}

// createAndStartServer creates a new server and starts it.
// It then waits until the instance is active.
func (vp *scalewayProvider) createAndStartServer(options providers.CreateInstanceOptions) (providers.ClusterInstance, error) {
	zeroInstance := providers.ClusterInstance{}

	// Validate input
	if options.TincIpv4 == "" {
		return zeroInstance, maskAny(fmt.Errorf("TincIpv4 is empty"))
	}

	// Fetch SSH keys
	sshKeys, err := providers.FetchSSHKeys(options.SSHKeyGithubAccount)
	if err != nil {
		return zeroInstance, maskAny(err)
	}

	// Create cloud-config
	// user-data
	ccOpts := options.NewCloudConfigOptions()
	ccOpts.PrivateIPv4 = "$private_ipv4"
	ccOpts.SshKeys = sshKeys
	_ /*userData*/, err = templates.Render(cloudConfigTemplate, ccOpts)
	if err != nil {
		return zeroInstance, maskAny(err)
	}

	var arch string
	switch options.TypeID[:2] {
	case "C1":
		arch = "arm"
	case "C2", "VC":
		arch = "x86_64"
	}

	// Find image
	imageID, err := vp.getImageID(options.ImageID, arch)
	if err != nil {
		vp.Logger.Errorf("getImageID failed: %#v", err)
		return zeroInstance, maskAny(err)
	}

	name := options.InstanceName
	image := &imageID
	dynamicIPRequired := !vp.NoIPv4
	//bootscript := ""

	volID := ""
	/*if options.TypeID != commercialTypeVC1 {
		volDef := api.ScalewayVolumeDefinition{
			Name:         vp.volumeName(name),
			Size:         volumeSize,
			Type:         volumeType,
			Organization: vp.organization,
		}
		volID, err = vp.client.PostVolume(volDef)
		if err != nil {
			vp.Logger.Errorf("PostVolume failed: %#v", err)
			return zeroInstance, maskAny(err)
		}
	}*/

	publicIPIdentifier := ""
	if options.RoleLoadBalancer && vp.ReserveLoadBalancerIP {
		ip, err := vp.getFreeIP()
		if err != nil {
			return zeroInstance, maskAny(err)
		}
		publicIPIdentifier = ip.ID
	}

	opts := api.ScalewayServerDefinition{
		Name:              name,
		Image:             image,
		Volumes:           map[string]string{},
		DynamicIPRequired: &dynamicIPRequired,
		//Bootscript:        &bootscript,
		Tags: []string{
			options.ClusterInfo.ID,
			options.TincIpv4,
		},
		Organization:   vp.Organization,
		CommercialType: options.TypeID,
		PublicIP:       publicIPIdentifier,
		EnableIPV6:     vp.EnableIPV6,
	}
	if volID != "" {
		opts.Volumes["0"] = volID
	}
	vp.Logger.Debugf("Creating server %s: %#v\n", name, opts)
	id, err := vp.client.PostServer(opts)
	if err != nil {
		vp.Logger.Errorf("PostServer failed: %#v", err)
		// Delete volume
		if volID != "" {
			if err := vp.client.DeleteVolume(volID); err != nil {
				vp.Logger.Errorf("DeleteVolume failed: %#v", err)
			}
		}
		return zeroInstance, maskAny(err)
	}

	// Start server
	if err := vp.client.PostServerAction(id, "poweron"); err != nil {
		vp.Logger.Errorf("poweron failed: %#v", err)
		return zeroInstance, maskAny(err)
	}

	// Wait until server starts
	server, err := vp.waitUntilServerActive(id, true)
	if err != nil {
		return zeroInstance, maskAny(err)
	}

	// Download & copy fleet,etcd
	instance := vp.clusterInstance(server, true)
	return instance, nil
}

// bootstrapServer copies etcd & fleet into the instances and runs the scaleway bootstrap script.
// It then reboots the instances and waits until it is active again.
func (vp *scalewayProvider) bootstrapServer(instance providers.ClusterInstance, options providers.CreateInstanceOptions, machineID string) error {
	if err := vp.copyEtcd(instance); err != nil {
		vp.Logger.Errorf("copy etcd failed: %#v", err)
		return maskAny(err)
	}
	if err := vp.copyFleet(instance); err != nil {
		vp.Logger.Errorf("copy fleet failed: %#v", err)
		return maskAny(err)
	}

	// Bootstrap
	bootstrapOptions := struct {
		ScalewayProviderConfig
		providers.CreateInstanceOptions
		MachineID string
	}{
		ScalewayProviderConfig: vp.ScalewayProviderConfig,
		CreateInstanceOptions:  options,
		MachineID:              machineID,
	}
	bootstrap, err := templates.Render(bootstrapTemplate, bootstrapOptions)
	if err != nil {
		return maskAny(err)
	}
	vp.Logger.Infof("Running bootstrap on %s. This may take a while...", instance.Name)
	if err := instance.RunScript(vp.Logger, bootstrap, "/root/pulcy-bootstrap.sh"); err != nil {
		// Failed expected because of a reboot
		vp.Logger.Debugf("bootstrap failed (expected): %#v", err)
	}

	vp.Logger.Infof("Done running bootstrap on %s, rebooting...", instance.Name)
	if err := vp.client.PostServerAction(instance.ID, "reboot"); err != nil {
		vp.Logger.Errorf("reboot failed: %#v", err)
		return maskAny(err)
	}
	time.Sleep(time.Second * 5)
	if _, err := vp.waitUntilServerActive(instance.ID, false); err != nil {
		return maskAny(err)
	}

	vp.Logger.Infof("Created server %s %s\n", instance.ID, instance.Name)

	return nil
}

func (vp *scalewayProvider) getFreeIP() (api.ScalewayIPDefinition, error) {
	ip, err := vp.client.NewIP()
	if err != nil {
		return api.ScalewayIPDefinition{}, maskAny(err)
	}
	return ip.IP, nil
}

func (vp *scalewayProvider) waitUntilServerActive(id string, bootstrapNeeded bool) (api.ScalewayServer, error) {
	currentState := ""
	for {
		server, err := vp.client.GetServer(id)
		if err != nil {
			return api.ScalewayServer{}, err
		}
		if server.State != currentState {
			currentState = server.State
			vp.Logger.Debugf("server state changed to '%s'", server.State)
		}
		if server.State == "running" {
			instance := vp.clusterInstance(*server, bootstrapNeeded)
			// Check SSH port state
			sshOpen, err := instance.IsSSHPortOpen(vp.Logger)
			if err != nil {
				vp.Logger.Errorf("Cannot check SSH port state: %#v", err)
			} else if sshOpen {
				// Attempt an SSH connection
				if _, err := instance.GetMachineID(vp.Logger); err == nil {
					// Success
					return *server, nil
				} else {
					vp.Logger.Debugf("get machine-id failed: %#v", err)
				}
			}
		}
		// Wait a while
		time.Sleep(time.Second * 5)
	}
}

// Create an entire cluster
func (vp *scalewayProvider) CreateCluster(log *logging.Logger, options providers.CreateClusterOptions, dnsProvider providers.DnsProvider) error {
	wg := sync.WaitGroup{}
	errors := make(chan error, options.InstanceCount)
	instanceDatas := make(chan instanceData, options.InstanceCount)
	for i := 1; i <= options.InstanceCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Duration((i - 1)) * time.Second * 10)
			isCore := (i <= 3)
			isLB := (i <= 2)
			instanceOptions, err := options.NewCreateInstanceOptions(isCore, isLB, i)
			if err != nil {
				errors <- maskAny(err)
				return
			}
			instance, err := vp.createInstance(log, instanceOptions, dnsProvider, nil)
			if err != nil {
				errors <- maskAny(err)
			} else {
				instanceDatas <- instanceData{
					CreateInstanceOptions: instanceOptions,
					ClusterInstance:       instance,
					FleetMetadata:         instanceOptions.CreateFleetMetadata(i),
				}
			}
		}(i)
	}
	wg.Wait()
	close(errors)
	close(instanceDatas)
	err := <-errors
	if err != nil {
		return maskAny(err)
	}

	instances := []instanceData{}
	instanceList := providers.ClusterInstanceList{}
	for data := range instanceDatas {
		instances = append(instances, data)
		instanceList = append(instanceList, data.ClusterInstance)
	}

	clusterMembers, err := instanceList.AsClusterMemberList(log, nil)
	if err != nil {
		return maskAny(err)
	}

	// Create tinc network config
	if instanceList.ReconfigureTincCluster(vp.Logger, instanceList); err != nil {
		return maskAny(err)
	}

	if err := vp.setupInstances(log, instances, clusterMembers); err != nil {
		return maskAny(err)
	}

	return nil
}

type instanceData struct {
	CreateInstanceOptions providers.CreateInstanceOptions
	ClusterInstance       providers.ClusterInstance
	FleetMetadata         string
}

func (vp *scalewayProvider) setupInstances(log *logging.Logger, instances []instanceData, clusterMembers providers.ClusterMemberList) error {
	wg := sync.WaitGroup{}
	errors := make(chan error, len(instances))
	for _, instance := range instances {
		wg.Add(1)
		go func(instance instanceData) {
			defer wg.Done()
			iso := providers.InitialSetupOptions{
				ClusterMembers: clusterMembers,
				FleetMetadata:  instance.FleetMetadata,
			}

			if err := instance.ClusterInstance.InitialSetup(log, instance.CreateInstanceOptions, iso, vp); err != nil {
				errors <- maskAny(err)
				return
			}
		}(instance)
	}
	wg.Wait()
	close(errors)
	err := <-errors
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (p *scalewayProvider) copyFleet(i providers.ClusterInstance) error {
	fleetFile := fmt.Sprintf("fleet-%s-linux-amd64", p.FleetVersion)
	url := fmt.Sprintf("https://github.com/coreos/fleet/releases/download/%s/%s.tar.gz", p.FleetVersion, fleetFile)
	p.Logger.Debugf("downloading %s", fleetFile)
	return maskAny(p.downloadAndCopyToInstance(url, i, "/tmp/fleet.tar.gz"))
}

func (p *scalewayProvider) copyEtcd(i providers.ClusterInstance) error {
	etcdFile := fmt.Sprintf("etcd-%s-linux-amd64", p.EtcdVersion)
	url := fmt.Sprintf("https://github.com/coreos/etcd/releases/download/%s/%s.tar.gz", p.EtcdVersion, etcdFile)
	p.Logger.Debugf("downloading %s", etcdFile)
	return maskAny(p.downloadAndCopyToInstance(url, i, "/tmp/etcd.tar.gz"))
}

func (p *scalewayProvider) downloadAndCopyToInstance(url string, i providers.ClusterInstance, instancePath string) error {
	localPath, err := p.dm.Download(url)
	if err != nil {
		return maskAny(err)
	}
	if err := i.CopyTo(p.Logger, localPath, instancePath); err != nil {
		return maskAny(err)
	}
	return nil
}

func (p *scalewayProvider) getImageID(imageName, arch string) (string, error) {
	images, err := p.client.GetImages()
	if err != nil {
		return "", maskAny(err)
	}
	for _, img := range *images {
		if img.Name == imageName {
			for _, v := range img.Versions {
				if v.ID == img.CurrentPublicVersion {
					for _, limg := range v.LocalImages {
						if limg.Arch == arch {
							return limg.ID, nil
						}
					}
				}
			}
		}
	}
	return "", maskAny(errgo.WithCausef(nil, NotFoundError, "Image '%s' not found", imageName))
}
