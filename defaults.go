package main

import (
	"os"
)

func defaultDomain() string {
	return os.Getenv("QUARK_DOMAIN")
}

func defaultPrivateRegistryUrl() string {
	return os.Getenv("QUARK_REGISTRY_URL")
}

func defaultPrivateRegistryUserName() string {
	return os.Getenv("QUARK_REGISTRY_USERNAME")
}

func defaultPrivateRegistryPassword() string {
	return os.Getenv("QUARK_REGISTRY_PASSWORD")
}

func defaultSshKeys() []string {
	return []string{os.Getenv("QUARK_SSH_KEY")}
}

func defaultVagrantFolder() string {
	return os.Getenv("QUARK_VAGRANT_FOLDER")
}

func defaultClusterRegion() string {
	return os.Getenv("QUARK_REGION")
}
