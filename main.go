package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"arvika.pulcy.com/iggi/droplets/providers"
	"arvika.pulcy.com/iggi/droplets/providers/cloudflare"
	"arvika.pulcy.com/iggi/droplets/providers/digitalocean"
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

	digitalOceanToken string
	cloudflareApiKey  string
	cloudflareEmail   string
)

func init() {
	cmdMain.PersistentFlags().StringVarP(&digitalOceanToken, "digitalocean-token", "t", "", "Digital Ocean token")
	cmdMain.PersistentFlags().StringVarP(&cloudflareApiKey, "cloudflare-apikey", "k", "", "Cloudflare API key")
	cmdMain.PersistentFlags().StringVarP(&cloudflareEmail, "cloudflare-email", "e", "", "Cloudflare email address")
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
	if digitalOceanToken == "" {
		Exitf("Please specify a token\n")
	}
	return digitalocean.NewProvider(digitalOceanToken)
}

func newDnsProvider() providers.DnsProvider {
	if cloudflareApiKey == "" {
		Exitf("Please specify a cloudflare-apikey\n")
	}
	if cloudflareEmail == "" {
		Exitf("Please specify a cloudflare-email\n")
	}
	return cloudflare.NewProvider(cloudflareApiKey, cloudflareEmail)
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
