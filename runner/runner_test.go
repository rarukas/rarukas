package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/rarukas/rarukas/server"
	"github.com/stretchr/testify/assert"
	"github.com/yamamoto-febc/go-arukas"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

type testArukasClient struct {
	readAppResult     *arukas.AppData
	readAppError      error
	createAppResult   *arukas.AppData
	createAppError    error
	deleteAppError    error
	readServiceResult *arukas.ServiceData
	readServiceError  error
	powerOnError      error
	waitForStateFunc  func(context.Context, string, string) error
}

func (c *testArukasClient) ReadApp(id string) (*arukas.AppData, error) {
	return c.readAppResult, c.readAppError
}

func (c *testArukasClient) CreateApp(param *arukas.RequestParam) (*arukas.AppData, error) {
	return c.createAppResult, c.createAppError
}

func (c *testArukasClient) DeleteApp(id string) error {
	return c.deleteAppError
}

func (c *testArukasClient) ReadService(id string) (*arukas.ServiceData, error) {
	return c.readServiceResult, c.readServiceError
}

func (c *testArukasClient) PowerOn(id string) error {
	return c.powerOnError
}

func (c *testArukasClient) WaitForState(ctx context.Context, serviceID string, status string) error {
	if c.waitForStateFunc == nil {
		return nil
	}
	return c.waitForStateFunc(ctx, serviceID, status)
}

func TestGenerateKeyPair(t *testing.T) {

	cfg := &Config{}
	r := &realRunner{cfg: cfg}

	public, private, err := r.generateKeyPair()
	assert.NoError(t, err)
	assert.NotEmpty(t, public)
	assert.NotEmpty(t, private)

	// can parse privateKey?
	signer, err := ssh.ParsePrivateKey([]byte(private))
	assert.NoError(t, err)
	assert.NotNil(t, signer)

	// can parse publicKey?
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(public))
	assert.NoError(t, err)
	assert.NotNil(t, publicKey)
}

func TestSetupKeyPair(t *testing.T) {

	t.Run("PublicKey is empty", func(t *testing.T) {
		cfg := &Config{
			PrivateKey: "xxx",
		}
		r := &realRunner{cfg: cfg}
		r.setupKeyPair()
		assert.NotEmpty(t, cfg.PublicKey)
		assert.Equal(t, "xxx", cfg.PrivateKey)
	})

	t.Run("PrivateKey is empty", func(t *testing.T) {
		cfg := &Config{
			PublicKey: "xxx",
		}
		r := &realRunner{cfg: cfg}
		r.setupKeyPair()
		assert.Equal(t, "xxx", cfg.PublicKey)
		assert.NotEmpty(t, cfg.PrivateKey)
	})

	t.Run("Both key is empty", func(t *testing.T) {
		cfg := &Config{}
		r := &realRunner{cfg: cfg}
		r.setupKeyPair()
		assert.NotEmpty(t, cfg.PublicKey)
		assert.NotEmpty(t, cfg.PrivateKey)
	})

}

var testArukasApp = &arukas.AppData{
	Data: &arukas.App{
		ID: "36EE605D-3351-4549-954E-D9A4FF768C75",
		Relationships: &arukas.AppRelationship{
			Service: &arukas.RelationshipData{
				Data: &arukas.Relationship{
					ID:   "36EE605D-3351-4549-954E-D9A4FF768C75",
					Type: arukas.TypeServices,
				},
			},
		},
	},
}
var testArukasService = &arukas.ServiceData{
	Data: &arukas.Service{
		ID: "36EE605D-3351-4549-954E-D9A4FF768C75",
		Attributes: &arukas.ServiceAttr{
			PortMappings: [][]*arukas.PortMapping{
				{
					{
						Host:          "example.arukascloud.io",
						Protocol:      "tcp",
						ContainerPort: 8080,
						ServicePort:   12345,
					},
					{
						Host:          "example.arukascloud.io",
						Protocol:      "tcp",
						ContainerPort: 2222,
						ServicePort:   22222,
					},
				},
			},
		},
	},
}

func TestStartArukas(t *testing.T) {

	ctx := context.Background()

	t.Run("Error when calling createApp API", func(t *testing.T) {
		expect := errors.New("test")
		r := &realRunner{
			cfg: &Config{
				ArukasClient: &testArukasClient{
					createAppError: expect,
				},
			},
		}
		r.setupKeyPair()

		host, port, err := r.startServer(ctx)
		assert.Empty(t, host)
		assert.Empty(t, port)
		assert.Error(t, err)
		assert.Equal(t, expect, err)
	})

	// powerOn
	t.Run("Error when calling powerOn API", func(t *testing.T) {
		expect := errors.New("test")
		r := &realRunner{
			cfg: &Config{
				ArukasClient: &testArukasClient{
					createAppResult: testArukasApp,
					powerOnError:    expect,
				},
			},
		}
		r.setupKeyPair()

		host, port, err := r.startServer(ctx)
		assert.Empty(t, host)
		assert.Empty(t, port)
		assert.Error(t, err)
		assert.Equal(t, expect, err)
	})

	// wait for running
	t.Run("Timeout occure when booting", func(t *testing.T) {
		r := &realRunner{
			cfg: &Config{
				ArukasClient: &testArukasClient{
					createAppResult: testArukasApp,
					waitForStateFunc: func(ctx context.Context, serviceID string, status string) error {
						<-ctx.Done()
						return ctx.Err()
					},
				},
				BootTimeout: time.Second,
			},
		}
		r.setupKeyPair()

		host, port, err := r.startServer(ctx)
		assert.Empty(t, host)
		assert.Empty(t, port)
		assert.Error(t, err)
	})

	t.Run("Should set AppData to current runner's field", func(t *testing.T) {
		r := &realRunner{
			cfg: &Config{
				ArukasClient: &testArukasClient{
					createAppResult:   testArukasApp,
					readServiceResult: testArukasService,
				},
				BootTimeout: 10 * time.Second,
			},
		}
		r.setupKeyPair()

		host, port, err := r.startServer(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, r.currentArukasApp)
		assert.Equal(t, "example.arukascloud.io", host)
		assert.Equal(t, 22222, port)
	})
}

func TestConnectToHost(t *testing.T) {

	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	log.SetOutput(ioutil.Discard)

	r := &realRunner{cfg: &Config{
		out:         stdOut,
		err:         stdErr,
		ExecTimeout: 10 * time.Second,
	}}
	r.setupKeyPair()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	errChan := make(chan error)

	go func() {
		errChan <- server.Start(ctx, &server.Config{
			PublicKey:       r.cfg.PublicKey,
			SSHServerAddr:   "127.0.0.1",
			SSHServerPort:   server.RarukasDefaultSSHPort,
			HealthCheckAddr: "127.0.0.1",
			HealthCheckPort: server.RarukasDefaultHTTPPort,
			Command:         "/bin/sh",
		})
	}()

	addr := fmt.Sprintf("127.0.0.1:%d", server.RarukasDefaultSSHPort)
	go func() {
		for {
			conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			conn.Close()
			errChan <- nil
			return
		}
	}()

	// waiting for ssh server available
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	t.Run("Connect to server", func(t *testing.T) {

		go func() {
			// connect
			client, session, err := r.newSSHSession("root", addr, []byte(r.cfg.PrivateKey))
			if err != nil {
				errChan <- err
				return
			}
			defer session.Close()
			defer client.Close()

			errChan <- nil
		}()

		select {
		case err := <-errChan:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	})

	t.Run("Execute command with writing to stdout", func(t *testing.T) {
		// write to stdout
		r.cfg.Commands = []string{"/bin/echo", "-n", "foobar"}
		go func() {
			errChan <- r.execCommand(ctx, "127.0.0.1", server.RarukasDefaultSSHPort)
		}()

		select {
		case err := <-errChan:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		assert.Equal(t, "foobar", stdOut.String())
	})

	t.Run("Execute command with writing to stderr", func(t *testing.T) {
		r.cfg.Commands = []string{"/bin/echo", "-n", "foobar", ">&2"}
		go func() {
			errChan <- r.execCommand(ctx, "127.0.0.1", server.RarukasDefaultSSHPort)
		}()

		select {
		case err := <-errChan:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		assert.Equal(t, "foobar", stdErr.String())
	})
}

func TestSCP(t *testing.T) {

	if _, err := os.Stat("tmp/"); err != nil {
		os.MkdirAll("tmp/tmp", 0755)  // nolint
		os.MkdirAll("tmp/work", 0755) // nolint
	}
	defer func() {
		os.RemoveAll("tmp/") // nolint
	}()

	r := &realRunner{cfg: &Config{
		ExecTimeout:   10 * time.Second,
		serverTmpDir:  "tmp/tmp",
		serverWorkDir: "tmp/work",
		CommandFile:   "test/dir1/test1.bash",
	}}
	r.setupKeyPair()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	errChan := make(chan error)

	go func() {
		errChan <- server.Start(ctx, &server.Config{
			PublicKey:       r.cfg.PublicKey,
			SSHServerAddr:   "127.0.0.1",
			SSHServerPort:   server.RarukasDefaultSSHPort + 1,
			HealthCheckAddr: "127.0.0.1",
			HealthCheckPort: server.RarukasDefaultHTTPPort + 1,
			Command:         "/bin/sh",
		})
	}()

	addr := fmt.Sprintf("127.0.0.1:%d", server.RarukasDefaultSSHPort+1)
	go func() {
		for {
			conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			conn.Close()
			errChan <- nil
			return
		}
	}()

	// waiting for ssh server available
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	t.Run("Upload command file to tmp dir", func(t *testing.T) {

		r.cfg.serverTmpDir = "tmp/tmp"
		r.cfg.CommandFile = "test/dir1/test1.bash"

		// test/dir1/test1.bash -> tmp/tmp/test1.bash
		err := r.uploadCommandFile(ctx, "127.0.0.1", server.RarukasDefaultSSHPort+1)
		assert.NoError(t, err)

		assert.FileExists(t, "tmp/tmp/test1.bash")

		// compare file contents
		src, err := ioutil.ReadFile("test/dir1/test1.bash")
		if err != nil {
			assert.Fail(t, err.Error())
		}

		dest, err := ioutil.ReadFile("tmp/tmp/test1.bash")
		if err != nil {
			assert.Fail(t, err.Error())
		}

		assert.Equal(t, src, dest)
	})

	t.Run("Upload source-dir to workdir", func(t *testing.T) {

		r.cfg.serverWorkDir = "tmp/work/"
		r.cfg.SyncDir = "test/dir1"
		r.cfg.CommandFile = ""

		err := r.uploadSourceDir(ctx, "127.0.0.1", server.RarukasDefaultSSHPort+1)
		assert.NoError(t, err)

		expects := []struct {
			source string
			dest   string
		}{
			{
				source: "test/dir1/test1.bash",
				dest:   "tmp/work/test1.bash",
			},
			{
				source: "test/dir1/test2.bash",
				dest:   "tmp/work/test2.bash",
			},
			{
				source: "test/dir1/dir2/test3.bash",
				dest:   "tmp/work/dir2/test3.bash",
			},
			{
				source: "test/dir1/dir2/test4.bash",
				dest:   "tmp/work/dir2/test4.bash",
			},
		}

		for _, expect := range expects {
			assert.FileExists(t, expect.source)
			assert.FileExists(t, expect.dest)
			// compare file contents
			src, err := ioutil.ReadFile(expect.source)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			dest, err := ioutil.ReadFile(expect.dest)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			assert.Equal(t, src, dest)
		}
	})

	t.Run("Download workdir to dest-dir", func(t *testing.T) {

		r.cfg.serverWorkDir = "tmp/work/"
		r.cfg.SyncDir = "test/dir1"
		r.cfg.CommandFile = ""

		// prepare workdir
		err := r.uploadSourceDir(ctx, "127.0.0.1", server.RarukasDefaultSSHPort+1)
		assert.NoError(t, err)

		// download
		err = r.downloadRemoteDir(ctx, "127.0.0.1", server.RarukasDefaultSSHPort+1)
		assert.NoError(t, err)

		expects := []struct {
			source string
			dest   string
		}{
			{
				source: "tmp/work/test1.bash",
				dest:   "test/dir1/test1.bash",
			},
			{
				source: "tmp/work/test2.bash",
				dest:   "test/dir1/test2.bash",
			},
			{
				source: "tmp/work/dir2/test3.bash",
				dest:   "test/dir1/dir2/test3.bash",
			},
			{
				source: "tmp/work/dir2/test4.bash",
				dest:   "test/dir1/dir2/test4.bash",
			},
		}

		for _, expect := range expects {
			assert.FileExists(t, expect.source)
			assert.FileExists(t, expect.dest)
			// compare file contents
			src, err := ioutil.ReadFile(expect.source)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			dest, err := ioutil.ReadFile(expect.dest)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			assert.Equal(t, src, dest)
		}
	})

}
