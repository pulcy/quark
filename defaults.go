package main

import (
	"os"
)

const (
	defaultInstanceCount  = 3
	defaultGluonImage     = "pulcy/gluon:20160214210824"
	defaultRebootStrategy = "etcd-lock"
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

func defaultSshKeyGithubAccount() string {
	return os.Getenv("QUARK_SSH_KEY_GITHUB_ACCOUNT")
}

func defaultVagrantFolder() string {
	return os.Getenv("QUARK_VAGRANT_FOLDER")
}
