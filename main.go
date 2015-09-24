package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"arvika.pulcy.com/iggi/droplets/providers"
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

	token string
)

func init() {
	cmdMain.PersistentFlags().StringVarP(&token, "token", "t", "", "Digital Ocean token")
}

func main() {
	cmdMain.Execute()
}

func showUsage(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

func loadDefaults(cmd *cobra.Command, args []string) {
	if token == "" {
		token = os.Getenv("DIGITALOCEAN_TOKEN")
	}
}

func newProvider() providers.CloudProvider {
	return digitalocean.NewProvider(token)
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
