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

package providers

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/op/go-logging"
)

// UpdateClusterMembers updates /etc/cluster-members on all instances of the cluster
func ReconfigureTincCluster(log *logging.Logger, info ClusterInfo, provider CloudProvider) error {
	// Load all instances
	instances, err := provider.GetInstances(info)
	if err != nil {
		return maskAny(err)
	}

	// Call reconfigure-tinc-host on all instances
	if instances.ReconfigureTincCluster(log); err != nil {
		return maskAny(err)
	}

	return nil
}

// UpdateClusterMembers updates /etc/cluster-members on all instances of the cluster
func (instances ClusterInstanceList) ReconfigureTincCluster(log *logging.Logger) error {
	// Now update all members in parallel
	vpnName := "pulcy"
	wg := sync.WaitGroup{}
	errorChannel := make(chan error, len(instances))
	for _, i := range instances {
		wg.Add(1)
		go func(i ClusterInstance) {
			defer wg.Done()
			if err := i.configureTincHost(log, vpnName, instances); err != nil {
				errorChannel <- maskAny(err)
			}
		}(i)
	}
	wg.Wait()
	close(errorChannel)
	for err := range errorChannel {
		return maskAny(err)
	}

	for _, x := range instances {
		if err := x.distributeTincHosts(log, vpnName, instances); err != nil {
			return maskAny(err)
		}
	}

	return nil
}

func (i ClusterInstance) configureTincHost(log *logging.Logger, vpnName string, instances ClusterInstanceList) error {
	connectTo := []string{}
	for _, x := range instances {
		if x.Name != i.Name {
			connectTo = append(connectTo, x.PrivateIpv4)
		}
	}
	if err := i.CreateTincConf(log, vpnName, connectTo); err != nil {
		return maskAny(err)
	}
	if err := i.CreateTincHostConf(log, vpnName); err != nil {
		return maskAny(err)
	}
	if err := i.CreateTincScripts(log, vpnName); err != nil {
		return maskAny(err)
	}
	if err := i.CreateTincService(log, vpnName); err != nil {
		return maskAny(err)
	}
	//Create key
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tincd -n %s -K", vpnName), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

func (i ClusterInstance) distributeTincHosts(log *logging.Logger, vpnName string, instances ClusterInstanceList) error {
	conf, err := i.GetTincHostConf(log, vpnName)
	if err != nil {
		return maskAny(err)
	}
	for _, x := range instances {
		if x.Name != i.Name {
			err := i.SetTincHostConf(log, vpnName, i.TincName(), conf)
			if err != nil {
				return maskAny(err)
			}
		}
	}
	return nil
}

func (i ClusterInstance) GetTincIP(log *logging.Logger) (string, error) {
	log.Debugf("Fetching tinc-ip on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "cat /etc/pulcy/tinc-ip", "", false)
	return id, maskAny(err)
}

func (i ClusterInstance) SetTincIP(log *logging.Logger) (string, error) {
	log.Debugf("Writing tinc-ip on %s", i.PublicIpv4)
	id, err := i.runRemoteCommand(log, "sudo tee /etc/pulcy/tinc-ip", i.TincIpv4, false)
	return id, maskAny(err)
}

// CreateTincConf creates a tinc.conf for the host of the given instance
func (i ClusterInstance) CreateTincConf(log *logging.Logger, vpnName string, connectTo []string) error {
	lines := []string{
		fmt.Sprintf("Name = %s", i.TincName()),
		"AddressFamily = ipv4",
		"Interface = tun0",
	}
	for _, name := range connectTo {
		lines = append(lines, fmt.Sprintf("ConnectTo = %s", name))
	}
	confDir := path.Join("/etc/tinc", vpnName)
	confPath := path.Join(confDir, "tinc.conf")
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo mkdir %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tee %s", confPath), strings.Join(lines, "\n"), false); err != nil {
		return maskAny(err)
	}
	return nil
}

// CreateTincConf creates a /etc/tinc/<vpnName>/hosts/<hostName> for the host of the given instance
func (i ClusterInstance) CreateTincHostConf(log *logging.Logger, vpnName string) error {
	lines := []string{
		fmt.Sprintf("Address = %s", i.PrivateIpv4),
		fmt.Sprintf("Subnet = %s/32", i.TincIpv4),
	}
	confDir := path.Join("/etc/tinc", vpnName, "hosts")
	confPath := path.Join(confDir, i.TincName())
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo mkdir %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tee %s", confPath), strings.Join(lines, "\n"), false); err != nil {
		return maskAny(err)
	}
	return nil
}

// CreateTincScripts creates a /etc/tinc/<vpnName>/tinc-up|down for the host of the given instance
func (i ClusterInstance) CreateTincScripts(log *logging.Logger, vpnName string) error {
	upLines := []string{
		"#!/bin/sh",
		fmt.Sprintf("ifconfig $INTERFACE %s netmask 255.255.255.0", i.TincIpv4),
	}
	downLines := []string{
		"#!/bin/sh",
		"ifconfig $INTERFACE down",
	}
	confDir := path.Join("/etc/tinc", vpnName)
	upPath := path.Join(confDir, "tinc-up")
	downPath := path.Join(confDir, "tinc-down")
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo mkdir %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tee %s", upPath), strings.Join(upLines, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tee %s", downPath), strings.Join(downLines, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo chmod 755 %s %s", upPath, downPath), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// GetTincHostConf reads a /etc/tinc/<vpnName>/hosts/<hostName> for the host of the given instance
func (i ClusterInstance) GetTincHostConf(log *logging.Logger, vpnName string) (string, error) {
	confDir := path.Join("/etc/tinc", vpnName, "hosts")
	confPath := path.Join(confDir, i.TincName())
	content, err := i.runRemoteCommand(log, "cat "+confPath, "", false)
	if err != nil {
		return "", maskAny(err)
	}
	return content, nil
}

// SetTincHostConf creates a /etc/tinc/<vpnName>/hosts/<hostName> from the given content
func (i ClusterInstance) SetTincHostConf(log *logging.Logger, vpnName, tincName, content string) error {
	confDir := path.Join("/etc/tinc", vpnName, "hosts")
	confPath := path.Join(confDir, tincName)
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo mkdir %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tee %s", confPath), content, false); err != nil {
		return maskAny(err)
	}
	return nil
}

// CreateTincConf creates a /etc/tinc/<vpnName>/hosts/<hostName> for the host of the given instance
func (i ClusterInstance) CreateTincService(log *logging.Logger, vpnName string) error {
	lines := []string{
		"[Unit]",
		fmt.Sprintf("Description=tinc for network %s", vpnName),
		"After=local-fs.target network-pre.target networking.service",
		"Before=network.target",
		"",
		"[Service]",
		"Type=simple",
		fmt.Sprintf("ExecStart=/usr/sbin/tincd -D -n %s", vpnName),
		fmt.Sprintf("ExecReload=/usr/sbin/tincd -n %s reload", vpnName),
		fmt.Sprintf("ExecStop=/usr/sbin/tincd -n %s stop", vpnName),
		"TimeoutStopSec=5",
		"Restart=always",
		"RestartSec=60",
		"",
		"[Install]",
		"WantedBy=multi-user.target",
	}
	confPath := "/etc/systemd/system/tinc.service"
	if _, err := i.runRemoteCommand(log, fmt.Sprintf("sudo tee %s", confPath), strings.Join(lines, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := i.runRemoteCommand(log, "sudo systemctl enable tinc.service", "", false); err != nil {
		return maskAny(err)
	}
	return nil
}
