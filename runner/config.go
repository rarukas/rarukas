package runner

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

// Config is configuration of rarukas cli runner
type Config struct {
	ArukasClient ArukasClient
	ArukasName   string
	ArukasPlan   string

	RarukasImageType string
	ArukasImageName  string

	PrivateKey string
	PublicKey  string

	CommandFile string
	SyncDir     string
	Commands    []string

	DownloadOnly bool
	UploadOnly   bool

	BootTimeout time.Duration
	ExecTimeout time.Duration

	serverTmpDir  string
	serverWorkDir string

	out io.Writer
	err io.Writer
	in  io.Reader
}

func (c *Config) hasCommandFile() bool {
	return c.CommandFile != ""
}

func (c *Config) hasSyncDir() bool {
	return c.SyncDir != ""
}

func (c *Config) syncDirExists() bool {
	if !c.hasSyncDir() {
		return false
	}
	_, e := os.Stat(c.SyncDir)
	return e == nil
}

func (c *Config) commandFileBase() string {
	if c.CommandFile == "" {
		return ""
	}
	return filepath.Base(c.CommandFile)
}
