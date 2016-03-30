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

package vultr

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/JamesClonk/vultr/lib"
	"github.com/op/go-logging"

	"github.com/pulcy/quark/providers"
	"github.com/pulcy/quark/templates"
)

const (
	fileMode            = os.FileMode(0775)
	cloudConfigTemplate = "templates/cloud-config.tmpl"
)

// Create a machine instance
func (vp *vultrProvider) CreateInstance(log *logging.Logger, options providers.CreateInstanceOptions, dnsProvider providers.DnsProvider) (providers.ClusterInstance, error) {
	// Create server
	id, err := vp.createServer(options)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	// Wait for the server to be active
	server, err := vp.waitUntilServerActive(id)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	publicIpv4 := server.MainIP
	publicIpv6 := ""
	if len(server.V6Networks) > 0 {
		publicIpv6 = server.V6Networks[0].MainIP
	}
	if err := providers.RegisterInstance(vp.Logger, dnsProvider, options, server.Name, options.RoleLoadBalancer, publicIpv4, publicIpv6); err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	vp.Logger.Info("Server '%s' is ready", server.Name)

	return vp.clusterInstance(server), nil
}

// Create a single server
func (vp *vultrProvider) createServer(options providers.CreateInstanceOptions) (string, error) {
	// Find SSH key ID
	var sshid string
	if len(options.SSHKeyNames) > 0 {
		var err error
		sshid, err = vp.findSSHKeyID(options.SSHKeyNames[0])
		if err != nil {
			return "", maskAny(err)
		}
	}
	// Fetch SSH keys
	sshKeys, err := providers.FetchSSHKeys(options.SSHKeyGithubAccount)
	if err != nil {
		return "", maskAny(err)
	}

	// Create cloud-config
	// user-data
	ccOpts := options.NewCloudConfigOptions()
	ccOpts.PrivateIPv4 = "$private_ipv4"
	ccOpts.SshKeys = sshKeys
	userData, err := templates.Render(cloudConfigTemplate, ccOpts)
	if err != nil {
		return "", maskAny(err)
	}

	name := options.InstanceName
	opts := &lib.ServerOptions{
		IPV6:              true,
		PrivateNetworking: true,
		SSHKey:            sshid,
		UserData:          userData,
	}
	regionID, err := strconv.Atoi(options.RegionID)
	if err != nil {
		return "", maskAny(err)
	}
	planID, err := strconv.Atoi(options.TypeID)
	if err != nil {
		return "", maskAny(err)
	}
	osID, err := strconv.Atoi(options.ImageID)
	if err != nil {
		return "", maskAny(err)
	}
	server, err := vp.client.CreateServer(name, regionID, planID, osID, opts)
	if err != nil {
		vp.Logger.Debug("CreateServer failed: %#v", err)
		return "", maskAny(err)
	}
	vp.Logger.Info("Created server %s %s\n", server.ID, server.Name)

	return server.ID, nil
}

func (vp *vultrProvider) waitUntilServerActive(id string) (lib.Server, error) {
	for {
		server, err := vp.client.GetServer(id)
		if err != nil {
			return lib.Server{}, err
		}
		if server.Status == "active" {
			// Attempt an SSH connection
			instance := vp.clusterInstance(server)
			if _, err := instance.GetMachineID(vp.Logger); err == nil {
				// Success
				return server, nil
			}
		}
		// Wait a while
		time.Sleep(time.Second * 5)
	}
}

// Create an entire cluster
func (vp *vultrProvider) CreateCluster(log *logging.Logger, options providers.CreateClusterOptions, dnsProvider providers.DnsProvider) error {
	wg := sync.WaitGroup{}
	errors := make(chan error, options.InstanceCount)
	instanceDatas := make(chan instanceData, options.InstanceCount)
	for i := 1; i <= options.InstanceCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Duration((i - 1)) * time.Second * 10)
			isCore := true
			isLB := true
			instanceOptions, err := options.NewCreateInstanceOptions(isCore, isLB, i)
			if err != nil {
				errors <- maskAny(err)
				return
			}
			instance, err := vp.CreateInstance(log, instanceOptions, dnsProvider)
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

func (vp *vultrProvider) setupInstances(log *logging.Logger, instances []instanceData, clusterMembers providers.ClusterMemberList) error {
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
			if err := instance.ClusterInstance.InitialSetup(log, instance.CreateInstanceOptions, iso); err != nil {
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
