// Code generated by go-bindata.
// sources:
// templates/cloud-config.tmpl
// templates/render.go
// templates/templates_bindata.go
// DO NOT EDIT!

package templates

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"os"
	"time"
	"io/ioutil"
	"path/filepath"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name string
	size int64
	mode os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _templatesCloudConfigTmpl = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x9c\x55\x61\x6f\xdb\x36\x10\xfd\xee\x5f\x71\x70\x0b\xe4\x4b\x69\x65\x6d\x80\x6e\x02\x84\xc1\x43\x0c\x2c\xc0\xd6\x19\x76\xb3\xa0\xd8\x86\x80\xa6\x4e\x36\x11\x8a\x54\xc9\x93\x53\xcf\xd3\x7f\xdf\x91\x72\x6c\xc7\x48\xe1\xa2\xc8\x17\xeb\x78\xef\xf1\xdd\xbd\x3b\xe6\x95\x32\xae\x2d\x85\x72\xb6\xd2\xcb\xc1\x40\x39\x8f\x2e\xe4\x03\x00\x24\x55\xbe\x8d\x3f\x00\x5e\xc1\x12\x2d\x7a\x49\x08\x12\x2c\x3e\x02\xb9\x07\xb4\x50\x39\x0f\x28\xd5\x0a\x5a\xab\x3f\xb7\x08\xca\xb4\x81\xd0\x43\xe5\x5d\x0d\x2b\xa2\x26\xe4\x59\x56\xea\xa0\xdc\x1a\xfd\x66\x14\x09\x47\xda\x65\x4c\xf0\x73\xd0\xff\x62\xf1\x2e\x91\xef\x13\x72\xd8\x6e\x3f\xb7\x8e\x2f\x19\x5d\x3f\xc5\x6e\xbd\xe9\xba\x9d\x86\xba\x35\xa4\x85\xc7\xa5\x76\x16\xa4\x2d\x77\x81\xa4\x1f\x4a\x6c\x8c\xdb\xd4\x68\x29\xb0\x42\x2c\x59\x22\xb4\x01\xe1\x75\xd3\x2e\x8c\x56\xf7\xba\x59\x5f\x25\x1e\x59\x32\x2f\xe9\x80\x0c\xd4\x9c\x2e\x5a\x6f\x42\x0e\xc3\xa8\x97\xe5\x6e\xb7\xa3\xa9\xd7\x6b\x2e\xf5\x66\xba\xbe\xea\xba\xfc\xed\xbb\xf7\x3f\x0d\x13\x52\x5b\x4d\x5a\x1a\x71\x60\x68\x10\xfd\x79\xfc\x8f\x97\xc3\x5d\x05\x46\x73\x7f\x2c\xb0\xfc\x85\xa3\x15\xd0\x0a\xc1\x55\x95\x56\xcc\x0a\x8d\xf3\x2c\x3d\x96\x15\xc3\x06\x97\x52\x6d\xfa\xe0\x13\xf8\x28\x04\x4a\x32\x07\xa3\x6b\x4d\xc4\xc5\xea\x0a\x36\xae\xf5\x20\x9b\x86\x8b\x95\x14\x3b\x54\x3a\x0c\xf6\x82\x62\x67\x90\x59\x39\xc2\xc4\x75\x22\xeb\x75\xbc\xdc\x80\xcb\x51\xfa\x4b\x75\xbf\x39\x89\x5d\x5d\x5e\xfe\x30\x3c\x66\xf8\xd6\x06\xbc\xf9\xda\xe1\xfb\x1d\x63\x65\x10\xa9\x1f\xb6\xde\x30\xa1\x9b\x9c\xcd\xeb\x93\x93\x7b\xa9\x09\x6c\x69\x99\xe6\x2e\x01\x14\x19\x08\x61\x05\xca\xd5\x35\x77\x2e\xc1\x6b\x24\x59\x4a\x92\xac\xa8\x1f\x95\x82\x2f\x9d\xa5\x5f\x5d\x17\xaf\x6a\x1b\x3e\xc6\x1c\x52\xb6\xc7\x85\x73\x24\x02\xc5\xe1\x5e\xf2\x08\x0e\xe3\x94\x0a\xe3\xd4\x43\xca\x65\xcb\x43\x2f\x4b\x80\x95\x35\xc3\xd2\x5a\x8c\x02\xfa\xb5\x56\x98\x4e\xe0\xe9\xfe\x1c\x02\x49\x4f\xcf\xd2\x93\xce\x6f\x4f\xdf\x48\x5f\x9e\xcf\xe6\xe5\xb4\x72\x61\x38\x9f\x7c\x7b\xc8\xb3\x6c\x09\xe5\xf0\xdf\x2e\x00\xf0\xd7\x2d\xeb\xff\x67\xff\x39\xae\x78\x3b\x8b\x92\x6b\x43\x7f\x72\x09\xe7\xce\xfb\xc0\x21\xfd\xe3\xa6\xc1\xc2\x59\x0c\x2b\x47\xfb\xe0\xc4\xae\xb5\x77\x36\x2e\x5a\xf1\x69\x3c\xbb\xbe\x9f\x8e\xe7\xf3\xe9\xaf\xb3\xf1\x7c\x52\xec\xd7\xf7\x13\x57\x31\x95\x21\x34\x2b\x2f\x03\xee\x16\xf8\x14\x3d\xff\x78\xfb\xe1\xc3\xe4\xb7\xfb\xe9\xe4\xf7\x17\x49\xe6\xd4\x5a\x8b\x66\x8a\xf5\x59\xaa\xbb\xc9\xf8\xcf\x49\x22\xb9\xfb\x63\x76\x7d\xa0\xb8\x43\xb9\xc6\x88\x7e\x74\xbe\x3c\xc6\x7e\x41\x35\x8f\xdd\x9c\x7a\x2c\xb2\x85\xb6\x19\x0f\x91\x50\x70\x51\x3f\x94\xda\x83\x68\x20\x5b\xb9\x1a\xb3\xf8\x18\xc6\xe3\x8b\x97\x91\xe2\x18\xea\x6b\x10\xd5\x09\x2e\x8b\x76\x7e\x05\x7c\x8c\xcd\xda\xe0\xd3\x77\xef\x0d\xaf\x80\x31\xfc\x18\xa6\x3e\xde\xd4\x72\xc9\x75\x7f\x07\x8b\x6f\x2d\x08\x11\x75\xad\x4f\x75\xe5\x59\x89\x81\xb4\x4d\x2f\x45\x06\x02\xe1\xc8\x81\xd7\xdb\x13\x63\xbb\xf3\x5a\x8a\xcc\x35\x94\xa8\x0f\x17\xc5\xe2\x21\x20\xb5\x0d\xfc\xbd\x47\x08\xde\xb5\xe4\x2b\xbf\x1d\xb5\x68\xf6\xce\xf2\xad\x2f\x0f\x44\xf7\x0c\xfc\x18\x1d\x4d\xb0\x68\x29\x83\x9e\x5b\xdf\xc1\x3e\x77\x86\xb5\xd4\x36\x8d\xfc\xe4\x8b\xa6\x62\x83\xe1\x30\xeb\x37\x96\x97\xc9\x98\xc3\xac\xdf\x49\x5e\x9e\xf2\x97\x4d\xd1\xff\x4f\xe1\x67\xc6\x8f\xb8\xae\x25\xd2\x60\xf0\x7f\x00\x00\x00\xff\xff\x14\x20\x04\x18\x22\x07\x00\x00")

func templatesCloudConfigTmplBytes() ([]byte, error) {
	return bindataRead(
		_templatesCloudConfigTmpl,
		"templates/cloud-config.tmpl",
	)
}

func templatesCloudConfigTmpl() (*asset, error) {
	bytes, err := templatesCloudConfigTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/cloud-config.tmpl", size: 1826, mode: os.FileMode(436), modTime: time.Unix(1443616132, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _templatesRenderGo = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x94\x52\x4d\x6f\xd4\x30\x10\x3d\xc7\xbf\x62\xc8\x01\x39\x28\x38\xea\x75\xa5\x1c\x8a\x44\x6f\xad\xf8\xba\x21\x84\xbc\x61\xb2\xb8\xdd\xd8\xc1\x1f\xfd\xd0\x6a\xff\x3b\x33\xb6\x77\xa1\x48\x1c\x7a\x8a\xe7\xbd\xc9\x9b\x37\xcf\x5e\xf5\x74\xa7\x77\x08\x11\x97\x75\xaf\x23\x06\x21\xcc\xb2\x3a\x1f\x41\x8a\xa6\xdd\x3e\x11\xd2\xd2\x21\x44\x3f\x39\x7b\xcf\xc7\x88\x8f\x71\x38\xb5\xb7\x82\x90\x9d\x89\x3f\xd3\x56\x4d\x6e\x19\x6e\xd3\x6d\x1a\xd0\xfb\x9d\x6b\x45\x27\xc4\xbd\xf6\xac\xb3\xe8\x70\x77\x69\x9f\x60\x84\x4c\xa9\x6b\xaa\xaf\x92\x9d\x64\x29\x89\xea\xb8\x7b\x26\x08\x3e\xa1\xfd\x81\x5e\x9e\x06\xdc\xe8\x05\x81\xa6\x1b\xbb\xeb\xc1\xad\xd1\x38\x1b\xc0\xd8\x88\x7e\xd6\x13\x1e\x8e\x1d\xc8\x13\x4b\x62\xce\x77\x70\x10\x8d\x0e\x01\x63\x06\x60\x33\xc2\x25\x57\xcf\x04\x3b\xd1\x98\x39\xd3\xaf\x46\xb0\x66\xcf\xff\x34\x1e\x63\xf2\x16\xda\xb6\x87\xea\x97\xed\x51\xeb\x91\x76\x1c\x06\x58\xb5\x0f\x7f\x72\x12\x0d\xef\x16\xa9\x80\x37\x27\x4c\x7d\x39\x93\x99\x18\xcf\xdd\xea\x06\x1f\xfe\x75\xc0\xdb\x5e\xeb\x95\x1d\x9e\xdb\xae\x0a\xc6\x76\x5a\x0c\x93\x5e\xb1\xdd\x40\x39\xf4\x8c\xfd\x4a\x2e\x32\x04\xf5\x3e\xd4\x47\x06\x7a\xf6\x98\x47\x66\x81\x20\xab\x34\x0d\xf9\x5e\x52\x18\xb3\x53\xf5\x81\x57\xa8\x79\xc9\x1c\x52\xf7\xd2\x28\x38\x89\x07\x6f\x22\xc2\x6c\xf6\x14\x87\x83\x6d\x9a\x67\xf4\xff\x99\x5f\x48\xde\xf1\x75\x7e\x4b\xea\x5d\x06\x0e\x24\xf4\x97\xaf\xf7\x8f\x38\xa5\x88\xb2\x74\x9f\xef\xf9\xe5\xd7\x54\xc9\xa2\xa3\x3e\x97\x45\xbb\x9e\x7f\x16\xc7\xfa\xc0\x4a\x9c\x32\xd4\x57\xd5\xd5\x2f\x8b\x07\x32\xf4\x2c\x59\xc9\x1e\xaa\x68\xf8\x7a\x01\x1b\xd8\xa3\x25\xf0\xed\xc5\x37\xd2\xfb\x1d\x00\x00\xff\xff\x1e\x0a\x18\x2c\x3c\x03\x00\x00")

func templatesRenderGoBytes() ([]byte, error) {
	return bindataRead(
		_templatesRenderGo,
		"templates/render.go",
	)
}

func templatesRenderGo() (*asset, error) {
	bytes, err := templatesRenderGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/render.go", size: 828, mode: os.FileMode(436), modTime: time.Unix(1443606994, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

var _templatesTemplates_bindataGo = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xac\x96\x4b\x8f\xdb\xc8\x11\xc7\xcf\xe2\xa7\xe0\x0e\xb0\x81\x04\x78\x35\x7c\x3f\x0c\xf8\xb2\x63\x07\xf0\x21\x5e\x20\xf1\x2d\x13\x18\xdd\xcd\xa6\x42\x44\x23\x2a\x94\xb4\x99\xf1\x62\xbf\x7b\xea\xd7\xd5\x9a\x9d\x49\xe0\xec\x25\x07\x4a\xcd\xea\xee\x7a\xfe\xeb\x5f\xbc\xbd\x4d\xef\xe6\xc1\xa7\x3b\x7f\xf0\x8b\x39\xfb\x21\xb5\x4f\xe9\x6e\xfe\xc1\x4e\x87\xc1\x9c\xcd\x36\x91\x03\xa7\xf9\xb2\x38\x7f\x7a\xcb\xfa\xec\x1f\x8e\x7b\x39\x77\xba\x75\xfb\xf9\x32\xfc\xe0\xe6\xc3\x38\xed\xb6\x67\x91\xbe\xde\x5e\xfc\x61\xf0\xcb\x76\x37\xbf\x16\x3f\xaf\xbe\x5c\x2d\xe8\x89\xf7\x3f\xa5\x9f\x7e\xfa\x9c\x7e\x78\xff\xf1\xf3\x77\x49\x72\x34\xee\x1f\x66\xe7\x7f\xbb\x97\x24\xd3\xc3\x71\x5e\xce\xe9\x3a\x59\xdd\xd8\x27\x91\xdc\xc8\xc2\xcd\x0f\xc7\xc5\x9f\x4e\xb7\xbb\xaf\xd3\x11\xc1\xf8\x70\xe6\x6f\x9a\xf9\x3d\x9d\x97\xe9\xb0\x0b\x07\xe7\xf0\x7b\x9e\x1e\xbc\x6e\xdf\x4e\xf3\xe5\x3c\xed\x79\x39\x9a\xf3\xdf\x6f\xc7\x69\xef\x59\xdc\x24\x9b\x24\x19\x2f\x07\x97\x46\xef\xfe\xec\xcd\xb0\x66\x91\xfe\xf5\x6f\x98\x7d\x93\x1e\xcc\x83\x4f\x55\xf5\x26\x5d\x5f\xa5\x7e\x59\xe6\x65\x93\xfe\x92\xac\x76\x5f\xc3\x5b\xfa\xf6\x5d\x8a\x57\xdb\x4f\xfe\x5f\x28\xf1\xcb\x3a\xb8\xcd\xfb\x8f\x97\x71\x94\x77\xd4\x6e\x36\xc9\x6a\x1a\xc3\x85\xef\xde\xa5\x87\x69\x8f\x8a\xd5\xe2\xcf\x97\xe5\xc0\xeb\x9b\x54\x42\xda\x7e\x40\xfb\xb8\xbe\x41\x51\xfa\xfd\x3f\xdf\xa6\xdf\xff\x7c\xa3\x9e\x04\x5b\xa2\xe3\xd7\x24\x59\xfd\x6c\x96\xd4\x5e\xc6\x54\xed\xa8\x91\x64\xf5\x45\xdd\x79\x97\x4e\xf3\xf6\x6e\x3e\x3e\xad\xff\x20\x67\xde\x88\x6f\x72\xcb\xed\x3f\x5c\x3d\xdd\xde\xed\xe7\x93\x5f\x4b\xf8\xff\x27\x7f\x50\xa3\xfa\xbf\xa1\x48\x0e\xaa\xdf\x51\x28\x6e\x6d\x7f\xc4\xf5\xf5\xe6\x0d\x27\x12\xd9\x3b\x3f\x1d\x7d\x6a\x4e\x27\x7f\x26\xe5\x17\x77\x46\x4b\x88\x2f\xd6\x43\xcc\x1c\xc6\x39\x4d\xe7\xd3\xf6\x8f\x52\xc3\x8f\xf2\xf2\x7c\x2f\x96\xf0\x2a\x7f\xa1\xe1\x45\x0d\x93\xd5\x69\xfa\xea\xd3\xe9\x70\x6e\xaa\x64\xf5\x40\x2b\x44\x5d\x7f\x92\x75\x90\x7c\x16\xd8\xa4\x60\x67\xcb\x0a\xf5\x01\x21\xeb\x71\xfa\x4f\x13\x9b\xf4\x93\x68\x5e\x6f\xa2\x6e\x4c\xc5\xe0\xc6\x69\x8b\x51\xb9\xfc\xed\xbb\x7f\x11\x47\xe4\x6e\x70\xe5\xf5\x55\x5c\xfc\x9f\x57\xf1\x55\xae\xbe\xf0\xfc\xb5\x02\xe2\xfa\x3d\x05\x04\x27\x3a\x9e\x03\xfd\x2f\x0d\x31\xfa\x6f\x2b\xf9\x78\x7a\x3f\x2d\xa2\xc2\xce\xf3\xfe\xe5\x6d\xb3\x3f\xfd\x4e\xe4\x4f\x27\x0d\xdc\x2f\xa3\x71\xfe\x97\x5f\x5f\xdc\x8e\x48\x00\xdc\x5f\x9e\xe9\xe0\x0e\xee\xb9\x0b\xd4\xf3\x59\x44\x82\x6e\x85\xc3\xfa\xe6\xfe\x31\x1f\xef\x1f\x3b\x7b\xff\x98\x75\xf2\x64\xf1\xe9\xef\x1f\x1b\x2f\xf2\x28\x1b\xe5\x4c\xef\xee\x1f\xeb\x5a\xe4\xb9\x3c\xf2\x3e\xc8\x9d\xb2\x91\xfb\xec\x0f\xf7\x8f\x5e\xce\xd7\x22\x6f\x65\xbf\x45\x87\xec\xfb\xea\xfe\xb1\x92\xff\x06\x7d\xdc\x95\x73\x5d\xa6\xba\xb3\x42\xd6\xb2\xef\xe4\x7c\x55\xca\xbb\xe8\x2f\xe4\x71\xb2\x3f\xa0\x57\xee\xb4\xf2\x6f\x65\xcf\x20\x13\x5f\xba\x46\xef\x1b\xf9\xaf\x7c\xb4\x2f\xf7\x3b\x23\xb6\xd1\x25\x77\x7a\x39\x5f\xcb\xe3\xf0\x51\xfe\x5b\xfe\xf1\x1f\xbf\xc4\x66\x83\x8d\x56\xee\x8b\x3e\x2f\x32\x23\x32\x27\xeb\x5c\x64\x5e\x7c\x6d\x91\xcb\xf9\x51\xf6\x06\xf1\xd7\xca\x53\x12\x8b\xd8\xea\x25\x86\x52\xce\x1b\xb1\x5d\x10\x4b\xad\x3a\xad\xec\x0d\xe8\xb1\x31\x1e\xf1\xb1\x14\x5f\x3a\xd1\x53\x78\xcd\x43\x56\x6a\x2e\x8b\x4a\xf3\x68\xc9\x2f\xba\x47\x95\xd7\x5e\xef\xe6\x85\xea\x6e\x25\xa6\xaa\xd7\xba\x04\x99\xf8\x5d\xc6\xbc\x59\xe4\xe2\xaf\x45\x57\xb4\x95\x61\x4b\xce\x0e\xa2\x3b\x23\x1f\xf2\x18\xab\xfa\x6d\xab\x7a\x9c\xd1\x3b\x1d\xb1\x65\x9a\x77\x2f\xe7\x6b\xb1\x97\xc9\x53\x58\xcd\x47\xd1\xa8\xcf\xb5\xe8\xad\x59\x9b\x58\x03\xf1\x21\x37\x5a\xef\x86\xdc\x57\x8a\x95\x4a\xf4\x0f\x46\x6b\x5c\xc9\x99\x51\xce\xb6\xa5\xda\x00\x3b\x4d\xa1\xf9\x24\x17\xd4\x81\xfc\xe6\xa5\xe6\x8e\xba\x82\x87\xab\x9f\x1d\xb9\x1d\xb5\xfe\x75\xcc\x8f\xab\xf5\x0c\xb5\xf0\x22\xb7\x62\xab\x1f\x14\x0f\xac\x4b\xd9\x37\x95\xd6\x83\xda\x9a\x5a\xb1\x93\x37\x2a\x0f\xf9\x6c\x35\x6e\xde\xc1\x7c\x65\x14\x0b\x9d\x53\x2c\x0f\x22\x6f\xc4\x97\x02\xfd\xe4\x26\xe6\x9c\x1c\x14\x85\xe6\x2d\xcb\x15\x33\x6d\xad\xd8\x22\x1e\xee\x93\xa3\xb1\x55\x5f\x42\xce\xa8\x85\x9c\xc9\xbd\xbe\x83\x19\x62\xc2\x77\xf0\x0b\xd6\xb1\x47\x4f\xa1\x83\x5a\xd1\x53\xd8\xcc\xc0\x01\x98\x17\x5f\xfa\x56\xeb\x13\x62\x27\x2e\xe4\xa2\x63\x90\x75\x63\xf5\xee\x58\x6b\x2d\xb0\x0d\xa6\xa8\x15\xeb\x51\x74\x7b\xf0\x6b\xd5\xb7\x80\xb3\x41\xf3\x5e\x8b\x9d\x5a\xe4\xd5\xa0\xb6\xa9\x29\x7d\xdb\x64\x9a\x83\x6b\x4f\xb7\xbd\xea\x01\xc7\xf8\x82\x5f\x01\x2b\xd4\xb9\xd1\x1a\xd3\x5f\xe4\x89\x9c\x85\xb3\xb5\xfa\x99\xd7\x5a\x7f\xb0\x08\xae\xc1\x3c\xfd\x43\xae\x9c\x53\x39\xf8\x1f\xcb\x88\x6d\x59\xf7\x8d\xde\xc3\x4e\xd6\xa8\x4d\x7c\x85\x37\x46\x39\xef\x33\xf5\x7f\xec\x14\xeb\x59\xa5\x75\x24\xbf\x59\xcc\x4f\x13\xe3\x72\x95\xe2\x86\x7c\x60\x0b\xae\x30\x31\x5f\x85\xe8\xe9\x6d\xe4\x13\xab\x7e\x81\x65\xfa\x1c\x1e\x29\x8d\xd6\x68\x64\xbf\x50\xfc\x36\xa2\xbf\xcf\x62\x3d\xe5\x71\x85\xda\x00\x0b\xe0\x23\x70\x07\x7d\xed\xb4\x4f\xa8\x33\x7c\x50\xe7\xea\x7b\x75\xdd\x1f\x23\x37\x44\x7c\x87\x07\xbf\xe8\x01\x7a\xd9\x69\x3d\x89\x11\x1e\x24\x0f\xe8\x1c\x7b\xed\x31\xf0\x37\x8a\xae\x9c\x1e\x2c\xb5\xe7\xa8\x17\xb8\x00\xd7\x81\x8b\xa2\x2e\x93\x6b\x9c\xe0\xcc\x0d\xea\x23\xfc\xc8\xd3\x5a\xbd\x13\xb8\xde\x29\x47\xf7\xbd\xf2\x1c\xb8\x02\xeb\xe4\x24\x8f\x1c\x04\xef\x53\x77\x38\x04\x8e\xa1\xe7\xe8\xa9\x22\xf2\x56\xc0\x63\xe4\x0c\x6a\xd2\xc3\x93\x4e\x71\x8c\x5f\xe4\xb8\x8b\x7d\x40\xef\xd2\xa3\xf4\x09\xf1\xc3\xbd\x06\x9e\x00\x8b\x56\x79\xce\x81\x05\xa7\xf5\x27\x26\xfa\x3a\xf0\x6d\xa9\x76\xe0\xc1\x50\xdf\x3c\xe6\xaf\x88\x33\x8c\x1e\x20\xc6\x4a\xeb\xce\x9a\x5e\x83\xeb\x5c\xe4\x28\xea\xc2\x3b\xbe\x50\x17\x6a\xd5\xc7\x98\xc0\x1e\x5c\x4e\x5f\xd0\xdb\xd8\xa1\x7e\x36\x62\xb7\xef\x74\xf6\x80\x45\xfa\x99\xba\xe5\xf1\x1c\x7c\x5d\x8d\x71\xf6\x14\x9a\x67\xe7\x63\x6f\xc7\x39\xc4\xdc\x21\xee\xde\xeb\x39\xf2\x4d\x8f\xc3\x2f\x61\x3e\x49\x3c\x25\xbd\xcf\x2c\x75\x5a\x27\xe6\x13\x73\x90\x3a\x51\x2f\xce\xc3\xb1\x43\x9c\x2f\x61\x46\x67\xca\x0b\x45\xac\x73\xe0\xdd\x56\xff\xc1\x1b\xd8\x67\xb6\x81\xf1\x80\xf9\x22\xf6\x4a\xec\x61\x6c\x07\xdd\xad\xfa\x0c\xc6\x98\xc5\x45\x1e\xe7\x7b\xa9\x73\x97\x5a\x83\x7d\xea\x1c\x66\xb6\x55\xfe\xe7\x3c\x39\xa1\x27\xb0\x65\x23\x3e\xca\xe6\x37\x8e\x63\x06\x80\x33\x30\x0e\x2f\xc1\x57\xc4\x09\xe7\x63\x17\x2e\x20\xa7\x60\x81\x78\x39\xdf\x46\x0c\x78\xa7\xf1\xc1\x65\x81\xc3\xd0\x9d\x6b\xdd\xf0\xb1\xac\xd4\x17\x6a\xca\xac\xe5\x7b\x24\x8f\x7d\x14\x7a\x7b\xd0\x5c\x90\x83\xe0\x6b\xae\xb5\x66\xce\xd9\xab\xff\xbd\xbe\x07\x1b\xad\xd6\x93\xfa\x80\x9f\x80\x07\xab\xbd\xc9\xf7\x06\x3c\x42\x0e\x8c\xd1\x58\xf8\xf6\xc0\x16\x35\xad\xae\x33\xa3\x57\xbd\xf4\x29\xf3\xa8\x1d\x74\x8e\xda\x4e\xf3\xce\x3e\x38\xe7\x3b\xab\xa5\x3e\x95\xe6\x84\xfc\x20\x47\x56\xe5\x8a\x4b\x38\x98\xef\x11\x7a\x99\x6f\x83\xc0\xb9\x85\xf6\x84\x8d\xdf\x4b\x59\xfc\xce\xe1\x9b\x80\x6f\x30\xb8\x07\xde\xa7\x17\xe0\x8e\xae\x8c\x33\x2e\x53\xfe\xc7\x3e\xdc\x6f\x23\xa7\xe6\x9d\xda\xa5\xd6\x01\x23\x6d\x9c\xd7\x85\xee\xc1\xa5\x60\x06\xbe\x09\xdf\x19\x7d\x9c\x81\x36\xc6\xe7\x95\x23\xda\x38\x57\xe1\x9c\x21\xf2\x76\xf8\x16\x6b\x75\xf6\x33\x87\xa8\x31\x7d\x47\x0d\x99\x1b\xe8\x07\x5f\xf4\x3e\xfc\x35\xc4\x7e\x20\xaf\xf0\x26\x18\x85\xe3\x43\x0d\x46\xc5\x79\xf8\x36\x8a\xf3\xd8\x0c\xda\x63\xac\x7d\xac\x0d\x1c\xcb\xcc\xa7\xb7\x98\x09\xa1\xdf\x6b\xe5\x54\xfa\x35\x60\xa0\x8b\xdf\x43\xad\x7e\x07\x80\x2d\xea\x15\xe6\x66\x19\xeb\x6e\x74\x66\x95\xf5\xbf\x03\x00\x00\xff\xff\xab\x3f\xba\x9c\x00\x10\x00\x00")

func templatesTemplates_bindataGoBytes() ([]byte, error) {
	return bindataRead(
		_templatesTemplates_bindataGo,
		"templates/templates_bindata.go",
	)
}

func templatesTemplates_bindataGo() (*asset, error) {
	bytes, err := templatesTemplates_bindataGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/templates_bindata.go", size: 12288, mode: os.FileMode(436), modTime: time.Unix(1443616381, 0)}
	a := &asset{bytes: bytes, info:  info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if (err != nil) {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"templates/cloud-config.tmpl": templatesCloudConfigTmpl,
	"templates/render.go": templatesRenderGo,
	"templates/templates_bindata.go": templatesTemplates_bindataGo,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"templates": &bintree{nil, map[string]*bintree{
		"cloud-config.tmpl": &bintree{templatesCloudConfigTmpl, map[string]*bintree{
		}},
		"render.go": &bintree{templatesRenderGo, map[string]*bintree{
		}},
		"templates_bindata.go": &bintree{templatesTemplates_bindataGo, map[string]*bintree{
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
        data, err := Asset(name)
        if err != nil {
                return err
        }
        info, err := AssetInfo(name)
        if err != nil {
                return err
        }
        err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
        if err != nil {
                return err
        }
        err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
        if err != nil {
                return err
        }
        err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
        if err != nil {
                return err
        }
        return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
        children, err := AssetDir(name)
        // File
        if err != nil {
                return RestoreAsset(dir, name)
        }
        // Dir
        for _, child := range children {
                err = RestoreAssets(dir, filepath.Join(name, child))
                if err != nil {
                        return err
                }
        }
        return nil
}

func _filePath(dir, name string) string {
        cannonicalName := strings.Replace(name, "\\", "/", -1)
        return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

