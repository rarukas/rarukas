package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rarukas/rarukas/runner"
	"github.com/rarukas/rarukas/version"
	"github.com/yamamoto-febc/go-arukas"
	"gopkg.in/urfave/cli.v2"
	"log"
	"time"
)

var (
	appName      = "rarukas"
	appUsage     = "CLI for running one-off commands on Arukas"
	appCopyright = "Copyright (C) 2018 Kazumichi Yamamoto."
)

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
	fmt.Fprintf( // nolint
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

	cfg.commands = c.Args().Slice()
	if len(cfg.commands) == 0 && cfg.commandFile == "" {
		return cli.ShowSubcommandHelp(c)
	}

	err := cfg.Validate()
	if err != nil {
		log.Printf("[ERROR] Initializing rarukas config failed\n%s", err)
		return err
	}

	// prepare SAKURA cloud API client
	arukasClient, err := arukas.NewClient(&arukas.ClientParam{
		Token:    cfg.accessToken,
		Secret:   cfg.accessTokenSecret,
		Trace:    cfg.traceMode,
		TraceOut: os.Stderr,
	})
	if err != nil {
		fmt.Printf("[ERROR] Initializing Arukas API Client failed\n%s", err)
		return err
	}

	runnerConfig := &runner.Config{
		ArukasClient:     arukasClient,
		ArukasName:       cfg.arukasName,
		ArukasPlan:       cfg.arukasPlan,
		RarukasImageType: cfg.rarukasImageType,
		ArukasImageName:  cfg.rarukasImageName,
		PublicKey:        cfg.publicKey,
		PrivateKey:       cfg.privateKey,
		CommandFile:      cfg.commandFile,
		SyncDir:          cfg.syncDir,
		UploadOnly:       cfg.uploadOnly,
		DownloadOnly:     cfg.downloadOnly,
		BootTimeout:      cfg.bootTimeout,
		ExecTimeout:      cfg.execTimeout,
		Commands:         cfg.commands,
	}

	// Setup signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-sigChan
		log.Printf("[INFO] Signal[%s] received. Shutting down...\n", signal.String())
		cancel()
	}()

	// Run
	if err := runner.Run(ctx, runnerConfig); err != nil {
		if err == ctx.Err() {
			time.Sleep(time.Second * 3) // sleep for shutting down goroutines
		} else {
			log.Fatal(err)
		}
	}

	log.Println("[INFO] Shutdown complete")
	return nil
}
