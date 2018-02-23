package runner

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/hnakamur/go-scp"
	"github.com/rarukas/rarukas/server"
	"github.com/yamamoto-febc/go-arukas"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Run starts rarukas-cli
func Run(ctx context.Context, cfg *Config) error {
	r := &realRunner{cfg: cfg}
	return r.run(ctx)
}

type realRunner struct {
	currentArukasApp *arukas.AppData
	cfg              *Config
}

func (r *realRunner) run(ctx context.Context) error {

	// setup key-pair
	if err := r.setupKeyPair(); err != nil {
		return nil
	}

	log.Print("[INFO] Starting rarukas-server on Arukas...")

	// start arukas container
	host, port, err := r.startServer(ctx)
	if err != nil {
		return err
	}
	// cleanup Arukas app after command execution
	defer r.cleanupServer()

	// sync (command-file + sync-dir)
	if r.cfg.hasCommandFile() {
		log.Print("[INFO] Uploading command-file to rarukas-server...")
		if err := r.uploadCommandFile(ctx, host, port); err != nil {
			return err
		}
	}
	if r.cfg.hasSyncDir() && r.cfg.syncDirExists() && !r.cfg.DownloadOnly {
		log.Print("[INFO] Uploading sync-dir to rarukas-server...")
		if err := r.uploadSourceDir(ctx, host, port); err != nil {
			return err
		}
	}

	log.Print("[INFO] Executing command on rarukas-server...")
	if err := r.execCommand(ctx, host, port); err != nil {
		return err
	}

	if r.cfg.hasSyncDir() && !r.cfg.UploadOnly {
		log.Print("[INFO] Downloading sync-dir from rarukas-server...")
		if err := r.downloadRemoteDir(ctx, host, port); err != nil {
			return err
		}
	}
	return nil
}

func (r *realRunner) setupKeyPair() error {
	if r.cfg.PublicKey == "" || r.cfg.PrivateKey == "" {
		publicKey, privateKey, err := r.generateKeyPair()
		if err != nil {
			return fmt.Errorf("[ERROR] generating key-pair failed: %s", err)
		}
		if r.cfg.PublicKey == "" {
			r.cfg.PublicKey = string(publicKey)
		}
		if r.cfg.PrivateKey == "" {
			r.cfg.PrivateKey = string(privateKey)
		}
	}
	return nil
}

func (r *realRunner) startServer(ctx context.Context) (string, int, error) {

	client := r.cfg.ArukasClient

	imageName := fmt.Sprintf("%s:%s", RarukasBaseImage, r.cfg.RarukasImageType)
	if r.cfg.ArukasImageName != "" {
		imageName = r.cfg.ArukasImageName
	}

	param := &arukas.RequestParam{
		Name:  r.cfg.ArukasName,
		Image: imageName,
		Plan:  r.cfg.ArukasPlan,
		Ports: []*arukas.Port{
			{
				Protocol: "tcp",
				Number:   server.RarukasDefaultHTTPPort,
			},
			{
				Protocol: "tcp",
				Number:   server.RarukasDefaultSSHPort,
			},
		},
		Environment: []*arukas.Env{
			{
				Key:   server.RarukasPublicKeyEnv,
				Value: r.cfg.PublicKey,
			},
			{
				Key:   server.RarukasCommandEnv,
				Value: "/bin/bash", // TODO make configurable??
			},
		},
		Instances: 1,
	}

	app, err := client.CreateApp(param)
	if err != nil {
		return "", 0, err
	}

	r.currentArukasApp = app
	serviceID := app.ServiceID()

	// power on
	if err = client.PowerOn(serviceID); err != nil {
		return "", 0, err
	}

	// Wait until container is running...
	ctx, cancel := context.WithTimeout(ctx, r.cfg.BootTimeout)
	defer cancel()

	errChan := make(chan error)
	go func() {
		errChan <- client.WaitForState(ctx, serviceID, arukas.StatusRunning)
	}()

	select {
	case err = <-errChan:
		if err != nil {
			return "", 0, err
		}
	case <-ctx.Done():
		r.cleanupServer()
		return "", 0, fmt.Errorf("Waiting for bootup of Arukas service timed out:\n\terror:%s", ctx.Err())
	}

	// get service port_mapping
	service, err := client.ReadService(serviceID)
	if err != nil {
		r.cleanupServer()
		return "", 0, err
	}

	portMapping := service.PortMapping()
	if len(portMapping) == 0 {
		r.cleanupServer()
		return "", 0, errors.New("Arukas service don't have port_mappings")
	}

	for _, pm := range portMapping {
		if pm.ContainerPort == server.RarukasDefaultSSHPort {
			return pm.Host, int(pm.ServicePort), nil
		}
	}

	r.cleanupServer()
	return "", 0, errors.New("Arukas service don't have SSH port_mapping")
}

func (r *realRunner) cleanupServer() {

	if r.currentArukasApp == nil {
		return
	}

	client := r.cfg.ArukasClient
	id := r.currentArukasApp.AppID()
	_, err := client.ReadApp(id)
	if err != nil {
		log.Printf("[ERROR] Cleanup failed: %s\n", err)
		return
	}
	if err := client.DeleteApp(id); err != nil {
		log.Printf("[ERROR] Cleanup failed: %s\n", err)
		return
	}
}

func (r *realRunner) execCommand(ctx context.Context, host string, port int) error {

	execCtx, cancel := context.WithTimeout(ctx, r.cfg.ExecTimeout)
	defer cancel()
	errChan := make(chan error)

	go func() {
		addr := fmt.Sprintf("%s:%d", host, port)
		client, session, err := r.newSSHSession("root", addr, []byte(r.cfg.PrivateKey))
		if err != nil {
			errChan <- err
			return
		}
		defer session.Close() // nolint -> return value not checked
		defer client.Close()  // nolint -> return value not checked

		cmd := strings.Join(r.cfg.Commands, " ")
		if r.cfg.hasCommandFile() {
			tmpDir := r.cfg.serverTmpDir
			if tmpDir == "" {
				tmpDir = RarukasServerTmpDir
			}
			if strings.HasSuffix(tmpDir, "/") {
				tmpDir = strings.TrimRight(tmpDir, "/")
			}
			cmd = fmt.Sprintf("/bin/bash %s/%s", tmpDir, r.cfg.commandFileBase())
		}

		out := r.cfg.out
		if out == nil {
			out = os.Stdout
		}
		errOut := r.cfg.err
		if errOut == nil {
			errOut = os.Stderr
		}
		in := r.cfg.in
		if in == nil {
			in = os.Stdin
		}

		session.Stdout = out
		session.Stderr = errOut
		session.Stdin = in
		errChan <- session.Run(cmd)
	}()

	select {
	case err := <-errChan:
		return err
	case <-execCtx.Done():
		return fmt.Errorf("[ERROR] Waiting for completion of command execution on Arukas timed out:\n\terror:%s", execCtx.Err())
	}
}

func (r *realRunner) uploadCommandFile(ctx context.Context, host string, port int) error {
	uploadCtx, cancel := context.WithTimeout(ctx, r.cfg.ExecTimeout)
	defer cancel()
	errChan := make(chan error)

	tmpDir := r.cfg.serverTmpDir
	if tmpDir == "" {
		tmpDir = RarukasServerTmpDir
	}
	if !strings.HasSuffix(tmpDir, "/") {
		tmpDir += "/"
	}

	go func() {
		errChan <- r.scpUpload(uploadCtx, host, port, r.cfg.CommandFile, tmpDir)
	}()

	select {
	case <-uploadCtx.Done():
		return uploadCtx.Err()
	case err := <-errChan:
		return err
	}
}

func (r *realRunner) uploadSourceDir(ctx context.Context, host string, port int) error {
	uploadCtx, cancel := context.WithTimeout(ctx, r.cfg.ExecTimeout)
	defer cancel()
	errChan := make(chan error)

	workDir := r.cfg.serverWorkDir
	if workDir == "" {
		workDir = RarukasServerWorkDir
	}
	if !strings.HasSuffix(workDir, "/") {
		workDir += "/"
	}

	go func() {
		errChan <- r.scpUpload(uploadCtx, host, port, r.cfg.SyncDir, workDir)
	}()

	select {
	case <-uploadCtx.Done():
		return uploadCtx.Err()
	case err := <-errChan:
		return err
	}
}

func (r *realRunner) downloadRemoteDir(ctx context.Context, host string, port int) error {
	downloadCtx, cancel := context.WithTimeout(ctx, r.cfg.ExecTimeout)
	defer cancel()
	errChan := make(chan error)

	workDir := r.cfg.serverWorkDir
	if workDir == "" {
		workDir = RarukasServerWorkDir
	}
	if !strings.HasSuffix(workDir, "/") {
		workDir += "/"
	}

	go func() {
		errChan <- r.scpDownload(downloadCtx, host, port, workDir, r.cfg.SyncDir)
	}()

	select {
	case <-downloadCtx.Done():
		return downloadCtx.Err()
	case err := <-errChan:
		return err
	}
}

func (r *realRunner) scpUpload(ctx context.Context, host string, port int, path string, destDir string) error {

	if !strings.HasSuffix(destDir, "/") {
		destDir += "/"
	}

	errChan := make(chan error)
	go func() {
		addr := fmt.Sprintf("%s:%d", host, port)
		client, err := r.openSSHConn("root", addr, []byte(r.cfg.PrivateKey))
		if err != nil {
			errChan <- err
			return
		}
		defer client.Close() // nolint -> return value not checked

		scpClient := scp.NewSCP(client)

		fi, err := os.Stat(path)
		if err != nil {
			errChan <- err
			return
		}

		// upload dir recursive
		if fi.IsDir() {

			// send files under srcDir
			entries, err := ioutil.ReadDir(path)
			if err != nil {
				errChan <- err
				return
			}

			for _, fi := range entries {
				switch {
				case fi.IsDir():
					if err := scpClient.SendDir(filepath.Join(path, fi.Name()), destDir, nil); err != nil {
						errChan <- err
						return
					}
				default:
					if err := scpClient.SendFile(filepath.Join(path, fi.Name()), destDir); err != nil {
						errChan <- err
						return
					}
				}
			}
			errChan <- nil
			return
		}

		// upload single file
		errChan <- scpClient.SendFile(path, destDir)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *realRunner) scpDownload(ctx context.Context, host string, port int, remoteDir, destDir string) error {

	if !strings.HasSuffix(remoteDir, "/") {
		remoteDir = remoteDir + "/"
	}

	if !strings.HasSuffix(destDir, "/") {
		destDir = destDir + "/"
	}

	errChan := make(chan error)
	go func() {
		var err error
		addr := fmt.Sprintf("%s:%d", host, port)

		client, err := r.openSSHConn("root", addr, []byte(r.cfg.PrivateKey))
		if err != nil {
			errChan <- err
			return
		}
		defer client.Close() // nolint -> return value not checked

		scpClient := scp.NewSCP(client)

		tmpDir, err := ioutil.TempDir("", "rarukas-download_")
		if err != nil {
			errChan <- err
			return
		}
		defer os.RemoveAll(tmpDir) // nolint

		// receive -> [destDir]/remoteDir.Base()
		if err = scpClient.ReceiveDir(remoteDir, tmpDir, nil); err != nil {
			errChan <- err
			return
		}
		tmpWorkDir := filepath.Join(tmpDir, filepath.Base(remoteDir))
		errChan <- r.mvDirFiles(tmpWorkDir, destDir)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (r *realRunner) mvDirFiles(srcDir, destDir string) error {
	// remove destDir/*
	if _, e := os.Stat(destDir); e == nil {

		files, err := ioutil.ReadDir(destDir)
		if err != nil {
			return err
		}

		for _, fi := range files {
			f := filepath.Join(destDir, fi.Name())
			if err := os.RemoveAll(f); err != nil {
				return err
			}
		}
	} else {
		if err := os.Mkdir(destDir, 0755); err != nil {
			return err
		}
	}

	// mv to under destDir/
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, fi := range files {
		src := filepath.Join(srcDir, fi.Name())
		dst := filepath.Join(destDir, fi.Name())
		if err = os.Rename(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func (r *realRunner) newSSHSession(user, host string, privateKey []byte) (*ssh.Client, *ssh.Session, error) {

	client, err := r.openSSHConn(user, host, privateKey)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close() // nolint return value not checked
		return nil, nil, err
	}

	return client, session, nil
}

func (r *realRunner) openSSHConn(user, host string, privateKey []byte) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", host, sshConfig)
}

func (r *realRunner) generateKeyPair() ([]byte, []byte, error) {
	reader := rand.Reader
	bitSize := 2048
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return nil, nil, err
	}

	// private key
	var privateKeyBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	privateKey := pem.EncodeToMemory(privateKeyBlock)

	// public key
	pub, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKey := ssh.MarshalAuthorizedKey(pub)

	return publicKey, privateKey, nil
}
