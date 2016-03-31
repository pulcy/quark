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

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"

	"github.com/pulcy/quark/providers"
	"github.com/pulcy/quark/providers/cloudflare"
	"github.com/pulcy/quark/providers/digitalocean"
	"github.com/pulcy/quark/providers/scaleway"
	"github.com/pulcy/quark/providers/vagrant"
	"github.com/pulcy/quark/providers/vultr"
)

var (
	projectVersion = "dev"
	projectBuild   = "dev"
	domain         = "pulcy.com"
)

const (
	projectName     = "quark"
	defaultLogLevel = "info"
)

var (
	cmdMain = &cobra.Command{
		Use:              projectName,
		Run:              showUsage,
		PersistentPreRun: loadDefaults,
	}

	provider             string
	digitalOceanToken    string
	cloudflareApiKey     string
	cloudflareEmail      string
	scalewayOrganization string
	scalewayToken        string
	vagrantFolder        string
	vultrApiKey          string
	logLevel             string

	log = logging.MustGetLogger(projectName)
)

func init() {
	logging.SetFormatter(logging.MustStringFormatter("[%{level:-5s}] %{message}"))
	cmdMain.PersistentFlags().StringVar(&logLevel, "log-level", defaultLogLevel, "Log level (debug|info|warning|error)")
	cmdMain.PersistentFlags().StringVarP(&provider, "provider", "p", "", "Provider used for creating clusters [digitalocean|scaleway|vagrant|vultr]")
	cmdMain.PersistentFlags().StringVarP(&digitalOceanToken, "digitalocean-token", "t", "", "Digital Ocean token")
	cmdMain.PersistentFlags().StringVarP(&cloudflareApiKey, "cloudflare-apikey", "k", "", "Cloudflare API key")
	cmdMain.PersistentFlags().StringVarP(&cloudflareEmail, "cloudflare-email", "e", "", "Cloudflare email address")
	cmdMain.PersistentFlags().StringVarP(&scalewayOrganization, "scaleway-organization", "", "", "Scaleway organization ID")
	cmdMain.PersistentFlags().StringVarP(&scalewayToken, "scaleway-token", "", "", "Scaleway token")
	cmdMain.PersistentFlags().StringVarP(&vagrantFolder, "vagrant-folder", "f", defaultVagrantFolder(), "Directory containing vagrant files")
	cmdMain.PersistentFlags().StringVarP(&vultrApiKey, "vultr-apikey", "", "", "Vultr API key")
}

func main() {
	cmdMain.Execute()
}

func showUsage(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

func loadDefaults(cmd *cobra.Command, args []string) {
	if digitalOceanToken == "" {
		digitalOceanToken = os.Getenv("DIGITALOCEAN_TOKEN")
	}
	if cloudflareApiKey == "" {
		cloudflareApiKey = os.Getenv("CLOUDFLARE_APIKEY")
	}
	if cloudflareEmail == "" {
		cloudflareEmail = os.Getenv("CLOUDFLARE_EMAIL")
	}
	if vultrApiKey == "" {
		vultrApiKey = os.Getenv("VULTR_APIKEY")
	}

	// Set loglevel
	level, err := logging.LogLevel(logLevel)
	if err != nil {
		Exitf("Invalid log-level '%s': %#v", logLevel, err)
	}
	logging.SetLevel(level, projectName)
}

func newProvider() providers.CloudProvider {
	switch provider {
	case "digitalocean":
		if digitalOceanToken == "" {
			Exitf("Please specify a token\n")
		}
		return digitalocean.NewProvider(log, digitalOceanToken)
	case "scaleway":
		if scalewayOrganization == "" || scalewayToken == "" {
			rc, err := scaleway.ReadRC()
			if err != nil {
				Exitf("Cannot read .scwrc: %#v\n", err)
			}
			scalewayOrganization = rc.Organization
			scalewayToken = rc.Token
		}
		if scalewayOrganization == "" {
			Exitf("Please specify a scaleway-organization\n")
		}
		if scalewayToken == "" {
			Exitf("Please specify a scaleway-token\n")
		}
		provider, err := scaleway.NewProvider(log, scalewayOrganization, scalewayToken)
		if err != nil {
			Exitf("NewProvider failed: %#v\n", err)
		}
		return provider
	case "vagrant":
		if vagrantFolder == "" {
			Exitf("Please specify a vagrant-folder\n")
		}
		return vagrant.NewProvider(log, vagrantFolder)
	case "vultr":
		if vultrApiKey == "" {
			Exitf("Please specify a vultr-apikey\n")
		}
		return vultr.NewProvider(log, vultrApiKey)
	default:
		Exitf("Unknown provider '%s'\n", provider)
		return nil
	}
}

func newDnsProvider() providers.DnsProvider {
	if cloudflareApiKey == "" {
		Exitf("Please specify a cloudflare-apikey\n")
	}
	if cloudflareEmail == "" {
		Exitf("Please specify a cloudflare-email\n")
	}
	return cloudflare.NewProvider(log, cloudflareApiKey, cloudflareEmail)
}

func confirm(question string) error {
	for {
		fmt.Printf("%s [yes|no]", question)
		bufStdin := bufio.NewReader(os.Stdin)
		line, _, err := bufStdin.ReadLine()
		if err != nil {
			return err
		}

		if string(line) == "yes" || string(line) == "y" {
			return nil
		}
		fmt.Println("Please enter 'yes' to confirm.")
	}
}

func Exitf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func Infof(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// clusterInfoFromArgs fills the given cluster info from a command line argument
func clusterInfoFromArgs(info *providers.ClusterInfo, args []string) {
	if len(args) == 1 && info.Name == "" {
		parts := strings.SplitN(args[0], ".", 2)
		info.Name = parts[0]
		info.Domain = parts[1]
	}
}

// clusterInstanceInfoFromArgs fills the given cluster info from a command line argument
func clusterInstanceInfoFromArgs(info *providers.ClusterInstanceInfo, args []string) {
	if len(args) == 1 && info.Prefix == "" && info.Name == "" {
		parts := strings.SplitN(args[0], ".", 3)
		info.Prefix = parts[0]
		info.Name = parts[1]
		info.Domain = parts[2]
	}
}
