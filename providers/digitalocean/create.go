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

package digitalocean

import (
	"errors"
	"sync"
	"time"

	"github.com/digitalocean/godo"

	"github.com/pulcy/quark/providers"
	"github.com/pulcy/quark/templates"
)

const (
	cloudConfigTemplate = "templates/cloud-config.tmpl"
)

func (dp *doProvider) CreateCluster(options providers.CreateClusterOptions, dnsProvider providers.DnsProvider) error {
	wg := sync.WaitGroup{}
	errors := make(chan error, options.InstanceCount)
	for i := 1; i <= options.InstanceCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			instanceOptions, err := options.NewCreateInstanceOptions()
			if err != nil {
				errors <- maskAny(err)
				return
			}
			_, err = dp.CreateInstance(instanceOptions, dnsProvider)
			if err != nil {
				errors <- maskAny(err)
			}
		}(i)
	}
	wg.Wait()
	close(errors)
	err := <-errors
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (dp *doProvider) CreateInstance(options providers.CreateInstanceOptions, dnsProvider providers.DnsProvider) (providers.ClusterInstance, error) {
	client := NewDOClient(dp.token)

	keys := []godo.DropletCreateSSHKey{}
	listedKeys, err := KeyList(client)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}
	for _, key := range options.SSHKeyNames {
		k := findKeyID(key, listedKeys)
		if k == nil {
			return providers.ClusterInstance{}, maskAny(errors.New("Key not found"))
		}
		keys = append(keys, godo.DropletCreateSSHKey{ID: k.ID})
	}

	opts := options.NewCloudConfigOptions()
	opts.PrivateIPv4 = "$private_ipv4"
	opts.PrivateClusterDevice = "eth1"

	cloudConfig, err := templates.Render(cloudConfigTemplate, opts)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	request := &godo.DropletCreateRequest{
		Name:              options.InstanceName,
		Region:            options.RegionID,
		Size:              options.TypeID,
		Image:             godo.DropletCreateImage{Slug: options.ImageID},
		SSHKeys:           keys,
		Backups:           false,
		IPv6:              true,
		PrivateNetworking: true,
		UserData:          cloudConfig,
	}

	// Create droplet
	dp.Logger.Info("Creating droplet: %s, %s, %s", request.Region, request.Size, options.ImageID)
	dp.Logger.Debug(cloudConfig)
	createDroplet, _, err := client.Droplets.Create(request)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	// Wait for active
	dp.Logger.Info("Waiting for droplet '%s'", createDroplet.Name)
	droplet, err := dp.waitUntilDropletActive(createDroplet.ID)
	if err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	publicIpv4 := getIpv4(*droplet, "public")
	publicIpv6 := getIpv6(*droplet, "public")
	if err := providers.RegisterInstance(dp.Logger, dnsProvider, options, createDroplet.Name, publicIpv4, publicIpv6); err != nil {
		return providers.ClusterInstance{}, maskAny(err)
	}

	dp.Logger.Info("Droplet '%s' is ready", createDroplet.Name)

	return dp.clusterInstance(*droplet), nil
}

func (dp *doProvider) waitUntilDropletActive(id int) (*godo.Droplet, error) {
	client := NewDOClient(dp.token)
	for {
		droplet, _, err := client.Droplets.Get(id)
		if err != nil {
			return nil, err
		}
		if droplet.Status != "new" {
			return droplet, nil
		}
		// Wait a while
		time.Sleep(time.Second * 5)
	}
}

func findKeyID(key string, listedKeys []godo.Key) *godo.Key {
	for _, k := range listedKeys {
		if k.Name == key {
			return &k
		}
	}
	return nil
}
