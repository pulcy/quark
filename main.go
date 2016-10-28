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
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clusterpkg "github.com/pulcy/quark/cluster"
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

	provider          string
	digitalOceanToken string
	cloudflareApiKey  string
	cloudflareEmail   string
	scalewayCfg       scaleway.ScalewayProviderConfig
	vagrantFolder     string
	vultrApiKey       string
	logLevel          string
	cluster           string
	vaultCfg          providers.VaultProviderConfig

	log = logging.MustGetLogger(projectName)
)

func init() {
	logging.SetFormatter(logging.MustStringFormatter("[%{level:-5s}] %{message}"))
	scalewayCfg = scaleway.NewConfig()
	cmdMain.PersistentFlags().StringVar(&logLevel, "log-level", defaultLogLevel, "Log level (debug|info|warning|error)")
	cmdMain.PersistentFlags().StringVarP(&provider, "provider", "p", "", "Provider used for creating clusters [digitalocean|scaleway|vagrant|vultr]")
	cmdMain.PersistentFlags().StringVarP(&cluster, "cluster", "c", "", "Path of the cluster template [<profile>@]path")

	// Digital ocean settings
	cmdMain.PersistentFlags().StringVarP(&digitalOceanToken, "digitalocean-token", "t", "", "Digital Ocean token")

	// Cloudflare settings
	cmdMain.PersistentFlags().StringVarP(&cloudflareApiKey, "cloudflare-apikey", "k", "", "Cloudflare API key")
	cmdMain.PersistentFlags().StringVarP(&cloudflareEmail, "cloudflare-email", "e", "", "Cloudflare email address")

	// Scaleway settings
	cmdMain.PersistentFlags().StringVarP(&scalewayCfg.Organization, "scaleway-organization", "", scalewayCfg.Organization, "Scaleway organization ID")
	cmdMain.PersistentFlags().StringVarP(&scalewayCfg.Token, "scaleway-token", "", scalewayCfg.Token, "Scaleway token")
	cmdMain.PersistentFlags().StringVarP(&scalewayCfg.Region, "scaleway-region", "", scalewayCfg.Region, "Scaleway region")
	cmdMain.PersistentFlags().BoolVar(&scalewayCfg.ReserveLoadBalancerIP, "scaleway-reserve-ip", scalewayCfg.ReserveLoadBalancerIP, "Use reserved IPv4 addresses for load-balancer instances")
	cmdMain.PersistentFlags().BoolVar(&scalewayCfg.EnableIPV6, "scaleway-ipv6", scalewayCfg.EnableIPV6, "Enabled IPv6 on all instances")
	cmdMain.PersistentFlags().BoolVar(&scalewayCfg.NoIPv4, "scaleway-no-ipv4", scalewayCfg.NoIPv4, "Do not add IPv4 addresses to new instances")

	// Vagrant settings
	cmdMain.PersistentFlags().StringVarP(&vagrantFolder, "vagrant-folder", "f", defaultVagrantFolder(), "Directory containing vagrant files")

	// Vultr settings
	cmdMain.PersistentFlags().StringVarP(&vultrApiKey, "vultr-apikey", "", "", "Vultr API key")

	// Vault settings
	vaultCfg.VaultCAPath = os.Getenv("VAULT_CAPATH")
	cmdMain.PersistentFlags().StringVar(&vaultCfg.VaultAddr, "vault-addr", defaultVaultAddr(), "URL of the vault (defaults to VAULT_ADDR environment variable)")
	cmdMain.PersistentFlags().StringVar(&vaultCfg.VaultCACert, "vault-cacert", defaultVaultCACert(), "Path to a PEM-encoded CA cert file to use to verify the Vault server SSL certificate")
	cmdMain.PersistentFlags().StringVar(&vaultCfg.VaultCAPath, "vault-capath", vaultCfg.VaultCAPath, "Path to a directory of PEM-encoded CA cert files to verify the Vault server SSL certificate")
	cmdMain.PersistentFlags().StringVarP(&vaultCfg.GithubToken, "github-token", "G", defaultGithubToken(), "Personal github token for administrator logins")
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
		if scalewayCfg.Organization == "" || scalewayCfg.Token == "" {
			rc, err := scaleway.ReadRC()
			if err != nil {
				Exitf("Cannot read .scwrc: %#v\n", err)
			}
			scalewayCfg.Organization = rc.Organization
			scalewayCfg.Token = rc.Token
		}
		if scalewayCfg.Organization == "" {
			Exitf("Please specify a scaleway-organization\n")
		}
		if scalewayCfg.Token == "" {
			Exitf("Please specify a scaleway-token\n")
		}
		provider, err := scaleway.NewProvider(log, scalewayCfg)
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

func newVaultProvider() providers.VaultProvider {
	provider, err := providers.NewVaultProvider(log, vaultCfg)
	if err != nil {
		Exitf("Failed to created vault provider: %#v\n", err)
	}
	return provider
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

// loadArgumentsFromCluster and uses its data to update the given flagset.
func loadArgumentsFromCluster(flagSet *pflag.FlagSet, requireProfile bool) {
	if cluster == "" {
		return
	}
	profile := ""
	parts := strings.Split(cluster, "@")
	if len(parts) == 2 {
		profile = parts[0]
		cluster = parts[1]
	} else if requireProfile {
		Exitf("No cluster profile specified (-c profile@cluster)")
	}
	clustersPath := os.Getenv("PULCY_CLUSTERS")
	if clustersPath == "" {
		clustersPath = "config/clusters"
	}
	path, err := resolvePath(cluster, clustersPath, ".hcl")
	if err != nil {
		Exitf("Cannot resolve cluster path: %#v", err)
	}
	c, err := clusterpkg.ParseClusterFromFile(path)
	if err != nil {
		Exitf("Cannot load cluster from path '%s': %#v", clustersPath, err)
	}
	values, err := c.ResolveProfile(profile)
	if err != nil {
		Exitf("Cannot resolve profile '%s' in cluster path '%s': %#v", profile, clustersPath, err)
	}
	flagSet.VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			value, ok := values[flag.Name]
			if ok {
				err := flagSet.Set(flag.Name, fmt.Sprintf("%v", value))
				if err != nil {
					Exitf("Error in option '%s': %#v\n", flag.Name, err)
				}
				log.Debugf("--%s=%v", flag.Name, value)
			}
		}
	})
}

// resolvePath tries to resolve a given path.
// 1) Try as real path
// 2) Try as filename relative to my process with given relative folder & extension
func resolvePath(path, altFolder, extension string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path not found, try locating it by name in a different folder
		var folder string
		if filepath.IsAbs(altFolder) {
			folder = altFolder
		} else {
			// altFolder is relative, assume it is relative to our executable
			exeFolder, err := osext.ExecutableFolder()
			if err != nil {
				return "", maskAny(err)
			}
			folder = filepath.Join(exeFolder, altFolder)
		}
		path = filepath.Join(folder, path) + extension
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Try without extensions
			path = filepath.Join(folder, path)
		}
	}
	return path, nil
}
