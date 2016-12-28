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
	"time"

	"github.com/op/go-logging"
	"golang.org/x/sync/errgroup"
)

// ReconfigureTincCluster creates the tinc configuration on all given instances.
func (instances ClusterInstanceList) ReconfigureTincCluster(log *logging.Logger, newInstances ClusterInstanceList) error {
	// Now update all members in parallel
	vpnName := "pulcy"
	wg := sync.WaitGroup{}
	errorChannel := make(chan error, len(instances))
	for _, i := range instances {
		wg.Add(1)
		go func(i ClusterInstance) {
			defer wg.Done()
			if newInstances.Contains(i) {
				if err := configureTincHost(log, i, vpnName, instances); err != nil {
					errorChannel <- maskAny(err)
				}
			} else {
				if err := reconfigureTincConf(log, i, vpnName, instances); err != nil {
					errorChannel <- maskAny(err)
				}
			}
		}(i)
	}
	wg.Wait()
	close(errorChannel)
	for err := range errorChannel {
		return maskAny(err)
	}

	g := errgroup.Group{}
	for _, i := range instances {
		i := i
		g.Go(func() error {
			if err := cleanupHostsFolder(log, i, vpnName); err != nil {
				return maskAny(err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}

	g = errgroup.Group{}
	for _, i := range instances {
		i := i
		g.Go(func() error {
			if err := distributeTincHosts(log, i, vpnName, instances); err != nil {
				return maskAny(err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return maskAny(err)
	}

	for _, i := range instances {
		// Restart tinc one after another
		if !newInstances.Contains(i) {
			if err := reloadTinc(log, i); err != nil {
				return maskAny(err)
			}
		} else {
			if err := restartTinc(log, i); err != nil {
				return maskAny(err)
			}
			time.Sleep(time.Second * 5)
		}
	}

	return nil
}

func cleanupHostsFolder(log *logging.Logger, i ClusterInstance, vpnName string) error {
	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	confDir := path.Join("/etc/tinc", vpnName)
	if _, err := s.Run(log, fmt.Sprintf("sudo cp -f %s/hosts/%s %s/self", confDir, tincName(i), confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo rm -f %s/hosts/*", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo cp -f %s/self %s/hosts/%s", confDir, confDir, tincName(i)), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

func configureTincHost(log *logging.Logger, i ClusterInstance, vpnName string, instances ClusterInstanceList) error {
	if err := reconfigureTincConf(log, i, vpnName, instances); err != nil {
		return maskAny(err)
	}
	if err := createTincHostsConf(log, i, vpnName); err != nil {
		return maskAny(err)
	}
	if err := createTincScripts(log, i, vpnName); err != nil {
		return maskAny(err)
	}
	if err := createTincService(log, i, vpnName); err != nil {
		return maskAny(err)
	}
	//Create key
	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()
	if _, err := s.Run(log, fmt.Sprintf("sudo tincd -n %s -K", vpnName), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

func reconfigureTincConf(log *logging.Logger, i ClusterInstance, vpnName string, instances ClusterInstanceList) error {
	log.Debugf("reconfigure tinc on %s", i)
	connectTo := []string{}
	for _, x := range instances {
		if x.Name != i.Name {
			connectTo = append(connectTo, tincName(x))
		}
	}
	if err := createTincService(log, i, vpnName); err != nil {
		return maskAny(err)
	}
	if err := createTincConf(log, i, vpnName, connectTo); err != nil {
		return maskAny(err)
	}
	return nil
}

func reloadTinc(log *logging.Logger, i ClusterInstance) error {
	// Reload config
	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()
	if _, err := s.Run(log, "sudo pkill -HUP tincd", "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

func restartTinc(log *logging.Logger, i ClusterInstance) error {
	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()
	log.Infof("Starting tinc on %s", i)
	if _, err := s.Run(log, "sudo systemctl restart tinc.service", "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// tincName creates the name of the instance in Tinc
func tincName(i ClusterInstance) string {
	return strings.Replace(strings.Replace(i.Name, ".", "_", -1), "-", "_", -1)
}

func distributeTincHosts(log *logging.Logger, i ClusterInstance, vpnName string, instances ClusterInstanceList) error {
	conf, err := getTincHostsConf(log, i, vpnName)
	if err != nil {
		return maskAny(err)
	}
	tincName := tincName(i)
	for _, x := range instances {
		if x.Name != i.Name {
			err := setTincHostsConf(log, x, vpnName, tincName, conf)
			if err != nil {
				return maskAny(err)
			}
		}
	}
	return nil
}

// createTincConf creates a tinc.conf for the host of the given instance
func createTincConf(log *logging.Logger, i ClusterInstance, vpnName string, connectTo []string) error {
	lines := []string{
		fmt.Sprintf("Name = %s", tincName(i)),
		"AddressFamily = ipv4",
		"Interface = tun0",
		"Cipher = aes-256-cbc",
		"Digest = sha256",
	}
	for _, name := range connectTo {
		lines = append(lines, fmt.Sprintf("ConnectTo = %s", name))
	}

	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	confDir := path.Join("/etc/tinc", vpnName)
	confPath := path.Join(confDir, "tinc.conf")
	if _, err := s.Run(log, fmt.Sprintf("sudo mkdir -p %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo tee %s", confPath), strings.Join(lines, "\n"), false); err != nil {
		return maskAny(err)
	}
	return nil
}

// createTincHostsConf creates a /etc/tinc/<vpnName>/hosts/<hostName> for the host of the given instance
func createTincHostsConf(log *logging.Logger, i ClusterInstance, vpnName string) error {
	address := i.PrivateIP
	lines := []string{
		fmt.Sprintf("Address = %s", address),
		fmt.Sprintf("Subnet = %s/32", i.ClusterIP),
	}
	if i.IsGateway {
		lines = append(lines, "Subnet = 0.0.0.0/0")
	}

	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	confDir := path.Join("/etc/tinc", vpnName, "hosts")
	confPath := path.Join(confDir, tincName(i))
	if _, err := s.Run(log, fmt.Sprintf("sudo mkdir -p %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo tee %s", confPath), strings.Join(lines, "\n"), false); err != nil {
		return maskAny(err)
	}
	return nil
}

// createTincScripts creates a /etc/tinc/<vpnName>/tinc-up|down for the host of the given instance
func createTincScripts(log *logging.Logger, i ClusterInstance, vpnName string) error {
	upLines := []string{
		"#!/bin/sh",
		fmt.Sprintf("ifconfig $INTERFACE %s netmask 255.255.255.0", i.ClusterIP),
	}
	if !i.IsGateway {
		upLines = append(upLines,
			"ORIGINAL_GATEWAY=$(ip route show | grep ^default | cut -d ' ' -f 2-5)",
			fmt.Sprintf("ip route replace %s $ORIGINAL_GATEWAY", i.PrivateNetwork.String()),
			"ip route replace 0.0.0.0/1 dev $INTERFACE",
			"ip route replace 128.0.0.0/1 dev $INTERFACE",
		)
	}

	downLines := []string{
		"#!/bin/sh",
		"ifconfig $INTERFACE down",
	}

	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	confDir := path.Join("/etc/tinc", vpnName)
	upPath := path.Join(confDir, "tinc-up")
	downPath := path.Join(confDir, "tinc-down")
	if _, err := s.Run(log, fmt.Sprintf("sudo mkdir -p %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo tee %s", upPath), strings.Join(upLines, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo tee %s", downPath), strings.Join(downLines, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo chmod 755 %s %s", upPath, downPath), "", false); err != nil {
		return maskAny(err)
	}
	return nil
}

// getTincHostsConf reads a /etc/tinc/<vpnName>/hosts/<hostName> for the host of the given instance
func getTincHostsConf(log *logging.Logger, i ClusterInstance, vpnName string) (string, error) {
	s, err := i.Connect()
	if err != nil {
		return "", maskAny(err)
	}
	defer s.Close()

	confDir := path.Join("/etc/tinc", vpnName, "hosts")
	confPath := path.Join(confDir, tincName(i))
	content, err := s.Run(log, "cat "+confPath, "", false)
	if err != nil {
		return "", maskAny(err)
	}
	return content, nil
}

// setTincHostsConf creates a /etc/tinc/<vpnName>/hosts/<hostName> from the given content
func setTincHostsConf(log *logging.Logger, i ClusterInstance, vpnName, tincName, content string) error {
	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	confDir := path.Join("/etc/tinc", vpnName, "hosts")
	confPath := path.Join(confDir, tincName)
	if _, err := s.Run(log, fmt.Sprintf("sudo mkdir -p %s", confDir), "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, fmt.Sprintf("sudo tee %s", confPath), content, false); err != nil {
		return maskAny(err)
	}
	return nil
}

// createTincService creates /etc/systemd/system/tinc.service on the given instance
func createTincService(log *logging.Logger, i ClusterInstance, vpnName string) error {
	lines := []string{
		"[Unit]",
		fmt.Sprintf("Description=tinc for network %s", vpnName),
		"After=local-fs.target network-pre.target networking.service",
		"Before=network.target",
		"",
		"[Service]",
		"Type=simple",
		fmt.Sprintf("ExecStart=/usr/sbin/tincd -D -n %s", vpnName),
		fmt.Sprintf("ExecReload=/usr/sbin/tincd -n %s -k HUP", vpnName),
		fmt.Sprintf("ExecStop=/usr/sbin/tincd -n %s -k", vpnName),
		"TimeoutStopSec=5",
		"Restart=always",
		"RestartSec=60",
		"",
		"[Install]",
		"WantedBy=multi-user.target",
	}

	s, err := i.Connect()
	if err != nil {
		return maskAny(err)
	}
	defer s.Close()

	confPath := "/etc/systemd/system/tinc.service"
	if _, err := s.Run(log, fmt.Sprintf("sudo tee %s", confPath), strings.Join(lines, "\n"), false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, "sudo systemctl daemon-reload", "", false); err != nil {
		return maskAny(err)
	}
	if _, err := s.Run(log, "sudo systemctl enable tinc.service", "", false); err != nil {
		return maskAny(err)
	}
	return nil
}
