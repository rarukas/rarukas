package main

import (
	"fmt"

	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
	"github.com/rarukas/rarukas/runner"
	"github.com/yamamoto-febc/go-arukas"
	"gopkg.in/urfave/cli.v2"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type config struct {
	accessToken       string
	accessTokenSecret string
	traceMode         bool

	publicKey  string
	privateKey string

	arukasName string
	arukasPlan string

	rarukasImageType string
	rarukasImageName string

	commands     []string
	commandFile  string
	syncDir      string
	downloadOnly bool
	uploadOnly   bool

	bootTimeout time.Duration
	execTimeout time.Duration
}

var cfg = &config{}

var cliFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "token",
		Usage:       "API Token of Arukas",
		EnvVars:     []string{"ARUKAS_JSON_API_TOKEN"},
		Destination: &cfg.accessToken,
	},
	&cli.StringFlag{
		Name:        "secret",
		Usage:       "API Secret of Arukas",
		EnvVars:     []string{"ARUKAS_JSON_API_SECRET"},
		Destination: &cfg.accessTokenSecret,
	},
	&cli.BoolFlag{
		Name:        "debug",
		Usage:       "Flag of debug-mode",
		EnvVars:     []string{"ARUKAS_DEBUG"},
		Destination: &cfg.traceMode,
		Value:       false,
		Hidden:      true,
	},
	&cli.StringFlag{
		Name:        "public-key",
		Usage:       "Public key for SSH auth. If empty, generate temporary key",
		EnvVars:     []string{"RARUKAS_PUBLIC_KEY"},
		Destination: &cfg.publicKey,
	},
	&cli.StringFlag{
		Name:        "private-key",
		Usage:       "Private key for SSH auth. If empty, generate temporary key",
		EnvVars:     []string{"RARUKAS_PRIVATE_KEY"},
		Destination: &cfg.privateKey,
	},
	&cli.StringFlag{
		Name:        "arukas-name",
		Usage:       "Name of Arukas app",
		EnvVars:     []string{"ARUKAS_NAME"},
		Value:       "rarukas-server",
		Destination: &cfg.arukasName,
	},
	&cli.StringFlag{
		Name: "arukas-plan",
		Usage: fmt.Sprintf("Plan of Arukas app [%s]",
			strings.Join(arukas.ValidPlans, "/"),
		),
		EnvVars:     []string{"ARUKAS_PLAN"},
		Value:       arukas.PlanFree,
		Destination: &cfg.arukasPlan,
	},
	&cli.StringFlag{
		Name:    "image-type",
		Aliases: []string{"type"},
		Usage: fmt.Sprintf(
			"OS Type of Rarukas server base image [%s]",
			strings.Join(runner.RarukasImageTypes, "/"),
		),
		EnvVars:     []string{"RARUKAS_IMAGE_TYPE"},
		Value:       "alpine",
		Destination: &cfg.rarukasImageType,
	},
	&cli.StringFlag{
		Name:        "image-name",
		Usage:       "Name of Rarukas server base image. It must exist in DockerHub. Ignore image-type if it was specified",
		EnvVars:     []string{"RARUKAS_IMAGE_NAME"},
		Destination: &cfg.rarukasImageName,
	},
	&cli.StringFlag{
		Name:        "command-file",
		Aliases:     []string{"c"},
		Usage:       "Script file to run on Arukas",
		EnvVars:     []string{"RARUKAS_COMMAND_FILE"},
		Destination: &cfg.commandFile,
	},
	&cli.StringFlag{
		Name:        "sync-dir",
		Usage:       "Directory to synchronize Arukas working directory",
		EnvVars:     []string{"RARUKAS_SYNC_DIR"},
		Destination: &cfg.syncDir,
	},
	&cli.BoolFlag{
		Name:        "download-only",
		Usage:       "Enable downloading only in synchronization with Arukas working directory",
		EnvVars:     []string{"RARUKAS_DOWNLOAD_ONLY"},
		Destination: &cfg.downloadOnly,
	},
	&cli.BoolFlag{
		Name:        "upload-only",
		Usage:       "Enable uploading only in synchronization with Arukas working directory",
		EnvVars:     []string{"RARUKAS_UPLOAD_ONLY"},
		Destination: &cfg.uploadOnly,
	},
	&cli.DurationFlag{
		Name:        "boot-timeout",
		Usage:       "Timeout duration when waiting for container be running",
		EnvVars:     []string{"RARUKAS_BOOT_TIMEOUT"},
		Destination: &cfg.bootTimeout,
		Value:       10 * time.Minute,
	},
	&cli.DurationFlag{
		Name:        "exec-timeout",
		Usage:       "Timeout duration when waiting for completion of command execution",
		EnvVars:     []string{"RARUKAS_EXEC_TIMEOUT"},
		Destination: &cfg.execTimeout,
		Value:       1 * time.Hour,
	},
}

func (c *config) Validate() error {
	var errs error

	validators := []func() error{
		// required
		func() error { return c.validateRequired("token", c.accessToken) },
		func() error { return c.validateRequired("secret", c.accessTokenSecret) },
		func() error { return c.validateRequired("arukas-name", c.accessTokenSecret) },
		func() error { return c.validateRequired("arukas-plan", c.accessTokenSecret) },
		func() error { return c.validateRequired("arukas-os-type", c.accessTokenSecret) },
		// valid word
		func() error {
			return c.validateStrInValues("arukas-plan", c.arukasPlan, arukas.ValidPlans...)
		},
		func() error {
			return c.validateStrInValues("image-type", c.rarukasImageType, runner.RarukasImageTypes...)
		},
		// file/dir
		func() error {
			return c.validateFilePath("command-file", c.commandFile)
		},
		func() error {
			return c.validateDirPath("sync-dir", c.syncDir)
		},
		func() error {
			if c.commandFile != "" && len(c.commands) > 0 {
				return errors.New("[Option] When --command-file is specified, no command-line argument can be specified")
			}
			return nil
		},
	}

	for _, v := range validators {
		err := v()
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func (c *config) validateRequired(name, v string) error {
	if v == "" {
		return fmt.Errorf("[Option] --%s is required", name)
	}
	return nil
}

func (c *config) validateStrInValues(name, v string, values ...string) error {
	if v == "" {
		return nil
	}

	exists := false
	for _, value := range values {
		if v == value {
			exists = true
			break
		}
	}
	if exists {
		return nil
	}
	return fmt.Errorf("[Option] --%s must be in [%s]", name, strings.Join(values, "/"))
}

func (c *config) validateFilePath(name, v string) error {

	if v == "" {
		return nil
	}

	path, err := homedir.Expand(v)
	if err != nil {
		return fmt.Errorf("[Option] --%s(%q) is invalid path", name, v)
	}

	cleaned := filepath.Clean(path)
	fi, err := os.Stat(cleaned)
	if err != nil {
		return fmt.Errorf("[Option] --%s(%q) is not exists", name, v)
	}

	if fi.IsDir() {
		return fmt.Errorf("[Option] --%s(%q) is directory", name, v)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("[Option] --%s(%q) is empty file", name, v)
	}
	return nil
}

func (c *config) validateDirPath(name, v string) error {
	if v == "" {
		return nil
	}

	path, err := homedir.Expand(v)
	if err != nil {
		return fmt.Errorf("[Option] --%s(%q) is invalid path", name, v)
	}

	cleaned := filepath.Clean(path)
	fi, err := os.Stat(cleaned)
	if err == nil && !fi.IsDir() {
		return fmt.Errorf("[Option] --%s(%q) is already exists file", name, v)
	}

	c.syncDir = cleaned
	return nil
}
