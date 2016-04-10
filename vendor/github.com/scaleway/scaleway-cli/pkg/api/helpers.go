// Copyright (C) 2015 Scaleway. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

package api

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/dustin/go-humanize"
	"github.com/moul/anonuuid"
	"github.com/scaleway/scaleway-cli/pkg/utils"
)

// ScalewayResolvedIdentifier represents a list of matching identifier for a specifier pattern
type ScalewayResolvedIdentifier struct {
	// Identifiers holds matching identifiers
	Identifiers ScalewayResolverResults

	// Needle is the criteria used to lookup identifiers
	Needle string
}

// ScalewayImageInterface is an interface to multiple Scaleway items
type ScalewayImageInterface struct {
	CreationDate time.Time
	Identifier   string
	Name         string
	Tag          string
	VirtualSize  float64
	Public       bool
	Type         string
	Organization string
	Archs        []string
	Region       string
}

// ResolveGateway tries to resolve a server public ip address, else returns the input string, i.e. IPv4, hostname
func ResolveGateway(api *ScalewayAPI, gateway string) (string, error) {
	if gateway == "" {
		return "", nil
	}

	// Parses optional type prefix, i.e: "server:name" -> "name"
	_, gateway = parseNeedle(gateway)

	servers, err := api.ResolveServer(gateway)
	if err != nil {
		return "", err
	}

	if len(servers) == 0 {
		return gateway, nil
	}

	if len(servers) > 1 {
		return "", showResolverResults(gateway, servers)
	}

	// if len(servers) == 1 {
	server, err := api.GetServer(servers[0].Identifier)
	if err != nil {
		return "", err
	}
	return server.PublicAddress.IP, nil
}

// CreateVolumeFromHumanSize creates a volume on the API with a human readable size
func CreateVolumeFromHumanSize(api *ScalewayAPI, size string) (*string, error) {
	bytes, err := humanize.ParseBytes(size)
	if err != nil {
		return nil, err
	}

	var newVolume ScalewayVolumeDefinition
	newVolume.Name = size
	newVolume.Size = bytes
	newVolume.Type = "l_ssd"

	volumeID, err := api.PostVolume(newVolume)
	if err != nil {
		return nil, err
	}

	return &volumeID, nil
}

// fillIdentifierCache fills the cache by fetching from the API
func fillIdentifierCache(api *ScalewayAPI, identifierType int) {
	log.Debugf("Filling the cache")
	var wg sync.WaitGroup
	wg.Add(5)
	go func() {
		if identifierType&(IdentifierUnknown|IdentifierServer) > 0 {
			api.GetServers(true, 0)
		}
		wg.Done()
	}()
	go func() {
		if identifierType&(IdentifierUnknown|IdentifierImage) > 0 {
			api.GetImages()
		}
		wg.Done()
	}()
	go func() {
		if identifierType&(IdentifierUnknown|IdentifierSnapshot) > 0 {
			api.GetSnapshots()
		}
		wg.Done()
	}()
	go func() {
		if identifierType&(IdentifierUnknown|IdentifierVolume) > 0 {
			api.GetVolumes()
		}
		wg.Done()
	}()
	go func() {
		if identifierType&(IdentifierUnknown|IdentifierBootscript) > 0 {
			api.GetBootscripts()
		}
		wg.Done()
	}()
	wg.Wait()
}

// GetIdentifier returns a an identifier if the resolved needles only match one element, else, it exists the program
func GetIdentifier(api *ScalewayAPI, needle string) (*ScalewayResolverResult, error) {
	idents := ResolveIdentifier(api, needle)

	if len(idents) == 1 {
		return &idents[0], nil
	}
	if len(idents) == 0 {
		return nil, fmt.Errorf("No such identifier: %s", needle)
	}

	sort.Sort(idents)
	for _, identifier := range idents {
		// FIXME: also print the name
		fmt.Fprintf(os.Stderr, "- %s\n", identifier.Identifier)
	}
	return nil, fmt.Errorf("Too many candidates for %s (%d)", needle, len(idents))
}

// ResolveIdentifier resolves needle provided by the user
func ResolveIdentifier(api *ScalewayAPI, needle string) ScalewayResolverResults {
	idents := api.Cache.LookUpIdentifiers(needle)
	if len(idents) > 0 {
		return idents
	}

	identifierType, _ := parseNeedle(needle)
	fillIdentifierCache(api, identifierType)

	idents = api.Cache.LookUpIdentifiers(needle)
	return idents
}

// ResolveIdentifiers resolves needles provided by the user
func ResolveIdentifiers(api *ScalewayAPI, needles []string, out chan ScalewayResolvedIdentifier) {
	// first attempt, only lookup from the cache
	var unresolved []string
	for _, needle := range needles {
		idents := api.Cache.LookUpIdentifiers(needle)
		if len(idents) == 0 {
			unresolved = append(unresolved, needle)
		} else {
			out <- ScalewayResolvedIdentifier{
				Identifiers: idents,
				Needle:      needle,
			}
		}
	}
	// fill the cache by fetching from the API and resolve missing identifiers
	if len(unresolved) > 0 {
		// compute identifierType:
		//   if identifierType is the same for every unresolved needle,
		//   we use it directly, else, we choose IdentifierUnknown to
		//   fulfill every types of cache
		identifierType, _ := parseNeedle(unresolved[0])
		for _, needle := range unresolved {
			newIdentifierType, _ := parseNeedle(needle)
			if identifierType != newIdentifierType {
				identifierType = IdentifierUnknown
				break
			}
		}

		// fill all the cache
		fillIdentifierCache(api, identifierType)

		// lookup again in the cache
		for _, needle := range unresolved {
			idents := api.Cache.LookUpIdentifiers(needle)
			out <- ScalewayResolvedIdentifier{
				Identifiers: idents,
				Needle:      needle,
			}
		}
	}

	close(out)
}

// InspectIdentifierResult is returned by `InspectIdentifiers` and contains the inspected `Object` with its `Type`
type InspectIdentifierResult struct {
	Type   int
	Object interface{}
}

// InspectIdentifiers inspects identifiers concurrently
func InspectIdentifiers(api *ScalewayAPI, ci chan ScalewayResolvedIdentifier, cj chan InspectIdentifierResult, arch string) {
	var wg sync.WaitGroup
	for {
		idents, ok := <-ci
		if !ok {
			break
		}
		idents.Identifiers = FilterImagesByArch(idents.Identifiers, arch)
		if len(idents.Identifiers) != 1 {
			if len(idents.Identifiers) == 0 {
				log.Errorf("Unable to resolve identifier %s", idents.Needle)
			} else {
				logrus.Fatal(showResolverResults(idents.Needle, idents.Identifiers))
			}
		} else {
			ident := idents.Identifiers[0]
			wg.Add(1)
			go func() {
				if ident.Type == IdentifierServer {
					server, err := api.GetServer(ident.Identifier)
					if err == nil {
						cj <- InspectIdentifierResult{
							Type:   ident.Type,
							Object: server,
						}
					}
				} else if ident.Type == IdentifierImage {
					image, err := api.GetImage(ident.Identifier)
					if err == nil {
						cj <- InspectIdentifierResult{
							Type:   ident.Type,
							Object: image,
						}
					}
				} else if ident.Type == IdentifierSnapshot {
					snap, err := api.GetSnapshot(ident.Identifier)
					if err == nil {
						cj <- InspectIdentifierResult{
							Type:   ident.Type,
							Object: snap,
						}
					}
				} else if ident.Type == IdentifierVolume {
					volume, err := api.GetVolume(ident.Identifier)
					if err == nil {
						cj <- InspectIdentifierResult{
							Type:   ident.Type,
							Object: volume,
						}
					}
				} else if ident.Type == IdentifierBootscript {
					bootscript, err := api.GetBootscript(ident.Identifier)
					if err == nil {
						cj <- InspectIdentifierResult{
							Type:   ident.Type,
							Object: bootscript,
						}
					}
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	close(cj)
}

// ConfigCreateServer represents the options sent to CreateServer and defining a server
type ConfigCreateServer struct {
	ImageName         string
	Name              string
	Bootscript        string
	Env               string
	AdditionalVolumes string
	DynamicIPRequired bool
	IP                string
	CommercialType    string
}

// CreateServer creates a server using API based on typical server fields
func CreateServer(api *ScalewayAPI, c *ConfigCreateServer) (string, error) {
	commercialType := os.Getenv("SCW_COMMERCIAL_TYPE")
	if commercialType == "" {
		commercialType = c.CommercialType
	}

	if c.Name == "" {
		c.Name = strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1)
	}

	var server ScalewayServerDefinition

	server.CommercialType = commercialType
	server.Volumes = make(map[string]string)
	server.DynamicIPRequired = &c.DynamicIPRequired
	if c.CommercialType == "" {
		return "", errors.New("You need to specify a commercial-type")
	}
	if c.IP != "" {
		if anonuuid.IsUUID(c.IP) == nil {
			server.PublicIP = c.IP
		} else {
			ips, err := api.GetIPS()
			if err != nil {
				return "", err
			}
			for _, ip := range ips.IPS {
				if ip.Address == c.IP {
					server.PublicIP = ip.ID
					break
				}
			}
			if server.PublicIP == "" {
				return "", fmt.Errorf("IP address %v not found", c.IP)
			}
		}
	}
	server.Tags = []string{}
	if c.Env != "" {
		server.Tags = strings.Split(c.Env, " ")
	}
	if c.AdditionalVolumes != "" {
		volumes := strings.Split(c.AdditionalVolumes, " ")
		for i := range volumes {
			volumeID, err := CreateVolumeFromHumanSize(api, volumes[i])
			if err != nil {
				return "", err
			}

			volumeIDx := fmt.Sprintf("%d", i+1)
			server.Volumes[volumeIDx] = *volumeID
		}
	}
	arch := os.Getenv("SCW_TARGET_ARCH")
	if arch == "" {
		switch server.CommercialType[:2] {
		case "C1":
			arch = "arm"
		case "C2", "VC":
			arch = "x86_64"
		}
	}
	region := os.Getenv("SCW_TARGET_REGION")
	if region == "" {
		region = "fr-1"
	}
	imageIdentifier := &ScalewayImageIdentifier{
		Arch:   arch,
		Region: region,
	}
	server.Name = c.Name
	inheritingVolume := false
	_, err := humanize.ParseBytes(c.ImageName)
	if err == nil {
		// Create a new root volume
		volumeID, errCreateVol := CreateVolumeFromHumanSize(api, c.ImageName)
		if errCreateVol != nil {
			return "", errCreateVol
		}
		server.Volumes["0"] = *volumeID
	} else {
		// Use an existing image
		inheritingVolume = true
		if anonuuid.IsUUID(c.ImageName) == nil {
			server.Image = &c.ImageName
		} else {
			imageIdentifier, err = api.GetImageID(c.ImageName, arch)
			if err != nil {
				return "", err
			}
			if imageIdentifier.Identifier != "" {
				server.Image = &imageIdentifier.Identifier
			} else {
				snapshotID, errGetSnapID := api.GetSnapshotID(c.ImageName)
				if errGetSnapID != nil {
					return "", errGetSnapID
				}
				snapshot, errGetSnap := api.GetSnapshot(snapshotID)
				if errGetSnap != nil {
					return "", errGetSnap
				}
				if snapshot.BaseVolume.Identifier == "" {
					return "", fmt.Errorf("snapshot %v does not have base volume", snapshot.Name)
				}
				server.Volumes["0"] = snapshot.BaseVolume.Identifier
			}
		}
	}
	if c.Bootscript != "" {
		bootscript, errGetBootScript := api.GetBootscriptID(c.Bootscript, imageIdentifier.Arch)
		if errGetBootScript != nil {
			return "", errGetBootScript
		}
		server.Bootscript = &bootscript
	}
	serverID, err := api.PostServer(server)
	if err != nil {
		return "", err
	}

	// For inherited volumes, we prefix the name with server hostname
	if inheritingVolume {
		createdServer, err := api.GetServer(serverID)
		if err != nil {
			return "", err
		}
		currentVolume := createdServer.Volumes["0"]

		var volumePayload ScalewayVolumePutDefinition
		newName := fmt.Sprintf("%s-%s", createdServer.Hostname, currentVolume.Name)
		volumePayload.Name = &newName
		volumePayload.CreationDate = &currentVolume.CreationDate
		volumePayload.Organization = &currentVolume.Organization
		volumePayload.Server.Identifier = &currentVolume.Server.Identifier
		volumePayload.Server.Name = &currentVolume.Server.Name
		volumePayload.Identifier = &currentVolume.Identifier
		volumePayload.Size = &currentVolume.Size
		volumePayload.ModificationDate = &currentVolume.ModificationDate
		volumePayload.ExportURI = &currentVolume.ExportURI
		volumePayload.VolumeType = &currentVolume.VolumeType

		err = api.PutVolume(currentVolume.Identifier, volumePayload)
		if err != nil {
			return "", err
		}
	}

	return serverID, nil
}

// WaitForServerState asks API in a loop until a server matches a wanted state
func WaitForServerState(api *ScalewayAPI, serverID string, targetState string) (*ScalewayServer, error) {
	var server *ScalewayServer
	var err error

	var currentState string

	for {
		server, err = api.GetServer(serverID)
		if err != nil {
			return nil, err
		}
		if currentState != server.State {
			log.Infof("Server changed state to '%s'", server.State)
			currentState = server.State
		}
		if server.State == targetState {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return server, nil
}

// WaitForServerReady wait for a server state to be running, then wait for the SSH port to be available
func WaitForServerReady(api *ScalewayAPI, serverID string, gateway string) (*ScalewayServer, error) {
	promise := make(chan bool)
	var server *ScalewayServer
	var err error
	var currentState string

	go func() {
		defer close(promise)

		for {
			server, err = api.GetServer(serverID)
			if err != nil {
				promise <- false
				return
			}
			if currentState != server.State {
				log.Infof("Server changed state to '%s'", server.State)
				currentState = server.State
			}
			if server.State == "running" {
				break
			}
			if server.State == "stopped" {
				err = fmt.Errorf("The server has been stopped")
				promise <- false
				return
			}
			time.Sleep(1 * time.Second)
		}

		if gateway == "" {
			log.Debugf("Waiting for server SSH port")
			dest := fmt.Sprintf("%s:22", server.PublicAddress.IP)
			err = utils.WaitForTCPPortOpen(dest)
			if err != nil {
				promise <- false
				return
			}
		} else {
			log.Debugf("Waiting for gateway SSH port")
			dest := fmt.Sprintf("%s:22", gateway)
			err = utils.WaitForTCPPortOpen(dest)
			if err != nil {
				promise <- false
				return
			}

			log.Debugf("Waiting 30 more seconds, for SSH to be ready")
			time.Sleep(30 * time.Second)
			// FIXME: check for SSH port through the gateway
		}
		promise <- true
	}()

	loop := 0
	for {
		select {
		case done := <-promise:
			utils.LogQuiet("\r \r")
			if !done {
				return nil, err
			}
			return server, nil
		case <-time.After(time.Millisecond * 100):
			utils.LogQuiet(fmt.Sprintf("\r%c\r", "-\\|/"[loop%4]))
			loop = loop + 1
			if loop == 5 {
				loop = 0
			}
		}
	}
}

// WaitForServerStopped wait for a server state to be stopped
func WaitForServerStopped(api *ScalewayAPI, serverID string) (*ScalewayServer, error) {
	server, err := WaitForServerState(api, serverID, "stopped")
	if err != nil {
		return nil, err
	}
	return server, nil
}

// ByCreationDate sorts images by CreationDate field
type ByCreationDate []ScalewayImageInterface

func (a ByCreationDate) Len() int           { return len(a) }
func (a ByCreationDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreationDate) Less(i, j int) bool { return a[j].CreationDate.Before(a[i].CreationDate) }

// StartServer start a server based on its needle, can optionaly block while server is booting
func StartServer(api *ScalewayAPI, needle string, wait bool) error {
	server, err := api.GetServerID(needle)
	if err != nil {
		return err
	}

	if err = api.PostServerAction(server, "poweron"); err != nil {
		if err.Error() == "server should be stopped" {
			return fmt.Errorf("server %s is already started: %v", server, err)
		}
	}

	if wait {
		_, err = WaitForServerReady(api, server, "")
		if err != nil {
			return fmt.Errorf("failed to wait for server %s to be ready, %v", needle, err)
		}
	}
	return nil
}

// StartServerOnce wraps StartServer for golang channel
func StartServerOnce(api *ScalewayAPI, needle string, wait bool, successChan chan bool, errChan chan error) {
	err := StartServer(api, needle, wait)

	if err != nil {
		errChan <- err
		return
	}

	fmt.Println(needle)
	successChan <- true
}

// DeleteServerSafe tries to delete a server using multiple ways
func (a *ScalewayAPI) DeleteServerSafe(serverID string) error {
	// FIXME: also delete attached volumes and ip address
	// FIXME: call delete and stop -t in parallel to speed up process
	err := a.DeleteServer(serverID)
	if err == nil {
		logrus.Infof("Server '%s' successfuly deleted", serverID)
		return nil
	}

	err = a.PostServerAction(serverID, "terminate")
	if err == nil {
		logrus.Infof("Server '%s' successfuly terminated", serverID)
		return nil
	}

	// FIXME: retry in a loop until timeout or Control+C
	logrus.Errorf("Failed to delete server %s", serverID)
	logrus.Errorf("Try to run 'scw rm -f %s' later", serverID)
	return err
}

// GetSSHFingerprintFromServer returns an array which containts ssh-host-fingerprints
func (a *ScalewayAPI) GetSSHFingerprintFromServer(serverID string) []string {
	ret := []string{}

	if value, err := a.GetUserdata(serverID, "ssh-host-fingerprints", false); err == nil {
		PublicKeys := strings.Split(string(*value), "\n")
		for i := range PublicKeys {
			if fingerprint, err := utils.SSHGetFingerprint([]byte(PublicKeys[i])); err == nil {
				ret = append(ret, fingerprint)
			}
		}
	}
	return ret
}
