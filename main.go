package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"

	"arvika.pulcy.com/pulcy/droplets/providers"
	"arvika.pulcy.com/pulcy/droplets/providers/cloudflare"
	"arvika.pulcy.com/pulcy/droplets/providers/digitalocean"
	"arvika.pulcy.com/pulcy/droplets/providers/vagrant"
	"arvika.pulcy.com/pulcy/droplets/providers/vultr"
)

var (
	projectVersion = "dev"
	projectBuild   = "dev"
	domain         = "pulcy.com"
)

var (
	cmdMain = &cobra.Command{
		Use:              "droplets",
		Run:              showUsage,
		PersistentPreRun: loadDefaults,
	}

	provider          string
	digitalOceanToken string
	cloudflareApiKey  string
	cloudflareEmail   string
	vagrantFolder     string
	vultrApiKey       string

	log = logging.MustGetLogger(cmdMain.Use)
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		Exitf("Cannot get current directory: %#v\n", err)
	}
	logging.SetFormatter(logging.MustStringFormatter("[%{level:-5s}] %{message}"))
	cmdMain.PersistentFlags().StringVarP(&provider, "provider", "p", "digitalocean", "Provider used for creating clusters [digitalocean|vagrant]")
	cmdMain.PersistentFlags().StringVarP(&digitalOceanToken, "digitalocean-token", "t", "", "Digital Ocean token")
	cmdMain.PersistentFlags().StringVarP(&cloudflareApiKey, "cloudflare-apikey", "k", "", "Cloudflare API key")
	cmdMain.PersistentFlags().StringVarP(&cloudflareEmail, "cloudflare-email", "e", "", "Cloudflare email address")
	cmdMain.PersistentFlags().StringVarP(&vagrantFolder, "vagrant-folder", "f", dir, "Directory containing vagrant files")
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
}

func newProvider() providers.CloudProvider {
	switch provider {
	case "digitalocean":
		if digitalOceanToken == "" {
			Exitf("Please specify a token\n")
		}
		return digitalocean.NewProvider(log, digitalOceanToken)
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
		fmt.Printf("%s ", question)
		bufStdin := bufio.NewReader(os.Stdin)
		line, _, err := bufStdin.ReadLine()
		if err != nil {
			return err
		}

		if string(line) == "yes" {
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

func def(envKey, defaultValue string) string {
	s := os.Getenv(envKey)
	if s == "" {
		s = defaultValue
	}
	return s
}
