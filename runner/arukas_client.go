package runner

import (
	"context"
	"github.com/yamamoto-febc/go-arukas"
)

const (
	// RarukasBaseImage is default image name for rarukas-server
	RarukasBaseImage = "rarukas/rarukas-server"
	// RarukasServerTmpDir is tmp directory path on rarukas-server
	RarukasServerTmpDir = "/tmp"
	// RarukasServerWorkDir is working directory path on rarukas-server
	RarukasServerWorkDir = "/workdir"
)

// RarukasImageTypes is valid types for rarukas-server
var RarukasImageTypes = []string{
	"alpine",
	"ansible",
	"centos",
	"debian",
	"golang",
	"node",
	"php",
	"python",
	"python2",
	"ruby",
	"sacloud",
	"ubuntu",
}

// ArukasClient is Arukas API Client interface
type ArukasClient interface {
	ReadApp(id string) (*arukas.AppData, error)
	CreateApp(param *arukas.RequestParam) (*arukas.AppData, error)
	DeleteApp(id string) error
	ReadService(id string) (*arukas.ServiceData, error)
	PowerOn(id string) error
	WaitForState(ctx context.Context, serviceID string, status string) error
}
