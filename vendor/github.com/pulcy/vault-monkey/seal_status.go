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
	"github.com/spf13/cobra"

	"github.com/pulcy/vault-monkey/service"
)

var (
	cmdSealStatus = &cobra.Command{
		Use:   "seal-status",
		Short: "Show the seal status of the vault.",
		Run:   cmdSealStatusRun,
	}
)

func init() {
	cmdMain.AddCommand(cmdSealStatus)
}

func cmdSealStatusRun(cmd *cobra.Command, args []string) {
	vs, err := service.NewVaultService(log, globalFlags.VaultServiceConfig)
	if err != nil {
		Exitf("Failed to create vault service: %#v", err)
	}
	if err := vs.SealStatus(); err != nil {
		Exitf("Failed to show seal-status of vault: %#v", err)
	}
}
