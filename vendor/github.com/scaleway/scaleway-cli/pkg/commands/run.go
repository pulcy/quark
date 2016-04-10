// Copyright (C) 2015 Scaleway. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/scaleway/scaleway-cli/pkg/api"
	"github.com/scaleway/scaleway-cli/pkg/config"
	"github.com/scaleway/scaleway-cli/pkg/utils"
)

// RunArgs are flags for the `Run` function
type RunArgs struct {
	Bootscript     string
	Command        []string
	Gateway        string
	Image          string
	Name           string
	IP             string
	Tags           []string
	Volumes        []string
	Userdata       string
	CommercialType string
	Timeout        int64
	AutoRemove     bool
	TmpSSHKey      bool
	ShowBoot       bool
	Detach         bool
	Attach         bool
}

// AddSSHKeyToTags adds the ssh key in the tags
func AddSSHKeyToTags(ctx CommandContext, tags *[]string, image string) error {
	home, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("unable to find your home %v", err)
	}
	idRsa := filepath.Join(home, ".ssh", "id_rsa")
	if _, errStat := os.Stat(idRsa); errStat != nil {
		if os.IsNotExist(errStat) {
			logrus.Warnln("Unable to find your ~/.ssh/id_rsa")
			logrus.Warnln("Run 'ssh-keygen -t rsa'")
			return nil
		}
	}
	idRsa = strings.Join([]string{idRsa, ".pub"}, "")
	data, err := ioutil.ReadFile(idRsa)
	if err != nil {
		return fmt.Errorf("failed to read %v", err)
	}
	data[7] = '_'
	for i := range data {
		if data[i] == ' ' {
			data = data[:i]
			break
		}
	}
	*tags = append(*tags, strings.Join([]string{"AUTHORIZED_KEY", string(data[:len(data)])}, "="))
	return nil
}

func addUserData(ctx CommandContext, userdatas []string, serverID string) {
	for i := range userdatas {
		keyValue := strings.Split(userdatas[i], "=")
		if len(keyValue) != 2 {
			logrus.Warn("Bad format: ", userdatas[i])
			continue
		}
		var data []byte
		var err error

		// Set userdata
		if keyValue[1][0] == '@' {
			data, err = ioutil.ReadFile(keyValue[1][1:])
			if err != nil {
				logrus.Warn("ReadFile: ", err)
				continue
			}
		} else {
			data = []byte(keyValue[1])
		}
		if err = ctx.API.PatchUserdata(serverID, keyValue[0], data, false); err != nil {
			logrus.Warn("PatchUserdata: ", err)
			continue
		}
	}
}

func runShowBoot(ctx CommandContext, args RunArgs, serverID string, closeTimeout chan struct{}, timeoutExit chan struct{}) error {
	// Attach to server serial
	logrus.Info("Attaching to server console ...")
	gottycli, done, err := utils.AttachToSerial(serverID, ctx.API.Token)
	if err != nil {
		close(closeTimeout)
		return fmt.Errorf("cannot attach to server serial: %v", err)
	}
	utils.Quiet(true)
	notif, gateway, err := waitSSHConnection(ctx, args, serverID)
	if err != nil {
		close(closeTimeout)
		gottycli.ExitLoop()
		<-done
		return err
	}
	select {
	case <-timeoutExit:
		gottycli.ExitLoop()
		<-done
		utils.Quiet(false)
		return fmt.Errorf("Operation timed out")
	case sshConnection := <-notif:
		close(closeTimeout)
		gottycli.ExitLoop()
		<-done
		utils.Quiet(false)
		if sshConnection.err != nil {
			return sshConnection.err
		}
		if fingerprints := ctx.API.GetSSHFingerprintFromServer(serverID); len(fingerprints) > 0 {
			for i := range fingerprints {
				fmt.Fprintf(ctx.Stdout, "%s\n", fingerprints[i])
			}
		}
		server := sshConnection.server
		logrus.Info("Connecting to server ...")
		if err = utils.SSHExec(server.PublicAddress.IP, server.PrivateIP, []string{}, false, gateway); err != nil {
			return fmt.Errorf("Connection to server failed: %v", err)
		}
	}
	return nil
}

// Run is the handler for 'scw run'
func Run(ctx CommandContext, args RunArgs) error {
	if args.Gateway == "" {
		args.Gateway = ctx.Getenv("SCW_GATEWAY")
	}

	if args.TmpSSHKey {
		err := AddSSHKeyToTags(ctx, &args.Tags, args.Image)
		if err != nil {
			return err
		}
	}
	env := strings.Join(args.Tags, " ")
	volume := strings.Join(args.Volumes, " ")

	// create IMAGE
	logrus.Info("Server creation ...")
	config := api.ConfigCreateServer{
		ImageName:         args.Image,
		Name:              args.Name,
		Bootscript:        args.Bootscript,
		Env:               env,
		AdditionalVolumes: volume,
		DynamicIPRequired: false,
		IP:                args.IP,
		CommercialType:    args.CommercialType,
	}
	if args.IP == "dynamic" || (args.IP == "" && args.Gateway == "") {
		config.DynamicIPRequired = true
		config.IP = ""
	} else if args.IP == "none" || args.IP == "no" || (args.IP == "" && args.Gateway != "") {
		config.IP = ""
	}
	serverID, err := api.CreateServer(ctx.API, &config)
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}
	logrus.Infof("Server created: %s", serverID)

	if args.AutoRemove {
		defer ctx.API.DeleteServerSafe(serverID)
	}

	// start SERVER
	logrus.Info("Server start requested ...")
	if err = api.StartServer(ctx.API, serverID, false); err != nil {
		return fmt.Errorf("failed to start server %s: %v", serverID, err)
	}
	logrus.Info("Server is starting, this may take up to a minute ...")

	if args.Userdata != "" {
		addUserData(ctx, strings.Split(args.Userdata, " "), serverID)
	}
	// Sync cache on disk
	ctx.API.Sync()

	if args.Detach {
		fmt.Fprintln(ctx.Stdout, serverID)
		return nil
	}

	closeTimeout := make(chan struct{})
	timeoutExit := make(chan struct{})

	if args.Timeout > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(args.Timeout) * time.Second):
				close(timeoutExit)
			case <-closeTimeout:
				break
			}
		}()
	}
	if args.ShowBoot {
		return runShowBoot(ctx, args, serverID, closeTimeout, timeoutExit)
	} else if args.Attach {
		// Attach to server serial
		logrus.Info("Attaching to server console ...")
		gottycli, done, err := utils.AttachToSerial(serverID, ctx.API.Token)
		close(closeTimeout)
		if err != nil {
			return fmt.Errorf("cannot attach to server serial: %v", err)
		}
		<-done
		gottycli.Close()
	} else {
		notif, gateway, err := waitSSHConnection(ctx, args, serverID)
		if err != nil {
			close(closeTimeout)
			return err
		}
		select {
		case <-timeoutExit:
			return fmt.Errorf("Operation timed out")
		case sshConnection := <-notif:
			close(closeTimeout)
			if sshConnection.err != nil {
				return sshConnection.err
			}
			if fingerprints := ctx.API.GetSSHFingerprintFromServer(serverID); len(fingerprints) > 0 {
				for i := range fingerprints {
					fmt.Fprintf(ctx.Stdout, "%s\n", fingerprints[i])
				}
			}
			server := sshConnection.server
			// exec -w SERVER COMMAND ARGS...
			if len(args.Command) < 1 {
				logrus.Info("Connecting to server ...")
				if err = utils.SSHExec(server.PublicAddress.IP, server.PrivateIP, []string{}, false, gateway); err != nil {
					return fmt.Errorf("Connection to server failed: %v", err)
				}
			} else {
				logrus.Infof("Executing command: %s ...", args.Command)
				if err = utils.SSHExec(server.PublicAddress.IP, server.PrivateIP, args.Command, false, gateway); err != nil {
					return fmt.Errorf("command execution failed: %v", err)
				}
				logrus.Info("Command successfuly executed")
			}
		}
	}
	return nil
}

type notifSSHConnection struct {
	server *api.ScalewayServer
	err    error
}

func waitSSHConnection(ctx CommandContext, args RunArgs, serverID string) (chan notifSSHConnection, string, error) {
	notif := make(chan notifSSHConnection)
	// Resolve gateway
	gateway, err := api.ResolveGateway(ctx.API, args.Gateway)
	if err != nil {
		return nil, "", fmt.Errorf("cannot resolve Gateway '%s': %v", args.Gateway, err)
	}

	// waiting for server to be ready
	logrus.Debug("Waiting for server to be ready")
	// We wait for 30 seconds, which is the minimal amount of time needed by a server to boot
	go func() {
		server, err := api.WaitForServerReady(ctx.API, serverID, gateway)
		if err != nil {
			notif <- notifSSHConnection{
				err: fmt.Errorf("cannot get access to server %s: %v", serverID, err),
			}
			return
		}
		logrus.Debugf("SSH server is available: %s:22", server.PublicAddress.IP)
		logrus.Info("Server is ready !")
		notif <- notifSSHConnection{
			server: server,
		}
	}()
	return notif, gateway, nil
}
