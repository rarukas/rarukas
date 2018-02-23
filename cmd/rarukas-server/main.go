package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/rarukas/rarukas/server"
	"github.com/rarukas/rarukas/version"
	"gopkg.in/urfave/cli.v2"
	"log"
	"time"
)

var (
	appName      = "rarukas-server"
	appUsage     = "A remote-shell-server running on Arukas"
	appCopyright = "Copyright (C) 2018 Kazumichi Yamamoto."
)

type config struct {
	publicKey       string
	command         string
	healthCheckAddr string
	healthCheckPort int
	sshServerAddr   string
	sshServerPort   int
}

var cfg = &config{}

var cliFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "public-key",
		EnvVars:     []string{server.RarukasPublicKeyEnv},
		Destination: &cfg.publicKey,
	},
	&cli.StringFlag{
		Name:        "command",
		EnvVars:     []string{server.RarukasCommandEnv},
		Value:       "/bin/sh",
		Destination: &cfg.command,
	},
	&cli.StringFlag{
		Name:        "health-check-addr",
		EnvVars:     []string{"RARUKAS_HEALTH_CHECK_ADDR"},
		Destination: &cfg.healthCheckAddr,
	},
	&cli.IntFlag{
		Name:        "health-check-port",
		EnvVars:     []string{"RARUKAS_HEALTH_CHECK_PORT"},
		Value:       server.RarukasDefaultHTTPPort,
		Destination: &cfg.healthCheckPort,
	},
	&cli.StringFlag{
		Name:        "ssh-server-addr",
		EnvVars:     []string{"RARUKAS_SSH_SERVER_ADDR"},
		Destination: &cfg.sshServerAddr,
	},
	&cli.IntFlag{
		Name:        "ssh-server-port",
		EnvVars:     []string{"RARUKAS_SSH_SERVER_PORT"},
		Value:       server.RarukasDefaultSSHPort,
		Destination: &cfg.sshServerPort,
	},
}

func (o *config) Validate() error {
	var err error

	if o.publicKey == "" {
		err = multierror.Append(err, errors.New("[Option] --public-key is required"))
	}
	if !(1 <= o.healthCheckPort && o.healthCheckPort <= 65535) {
		err = multierror.Append(err, errors.New("[Option] --health-check-port is invalid"))
	}
	if !(1 <= o.sshServerPort && o.sshServerPort <= 65535) {
		err = multierror.Append(err, errors.New("[Option] --ssh-server-port is invalid"))
	}
	return err
}

func main() {

	log.SetFlags(0)

	app := &cli.App{
		Name:                  appName,
		Usage:                 appUsage,
		HelpName:              appName,
		Copyright:             appCopyright,
		EnableShellCompletion: true,
		Version:               version.FullVersion(),
		CommandNotFound:       cmdNotFound,
		Flags:                 cliFlags,
		Action:                cmdMain,
	}
	cli.InitCompletionFlag.Hidden = true

	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func cmdNotFound(c *cli.Context, command string) {
	fmt.Fprintf(
		os.Stderr,
		"%s: '%s' is not a %s command. See '%s --help'\n",
		c.App.Name,
		command,
		c.App.Name,
		c.App.Name,
	)
	os.Exit(1)
}

func cmdMain(c *cli.Context) error {

	err := cfg.Validate()
	if err != nil {
		return err
	}

	log.Println("[INFO] Start rarukas-server")

	serverConfig := &server.Config{
		PublicKey:       cfg.publicKey,
		Command:         cfg.command,
		HealthCheckAddr: cfg.healthCheckAddr,
		HealthCheckPort: cfg.healthCheckPort,
		SSHServerAddr:   cfg.sshServerAddr,
		SSHServerPort:   cfg.sshServerPort,
	}

	// Setup signal handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-sigChan
		log.Printf("[INFO] Signal[%s] received. Shutting down...\n", signal.String())
		cancel()
	}()

	if err := server.Start(ctx, serverConfig); err != nil {
		if err == ctx.Err() {
			time.Sleep(time.Second * 3) // sleep for shutting down goroutines
		} else {
			log.Fatal(err)
		}
	}

	log.Println("[INFO] Shutdown complete")
	return nil
}
