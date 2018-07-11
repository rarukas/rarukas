// +build !windows

package server

import (
	"context"
	"net/http"

	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// Config is configuration of rarukas-server
type Config struct {
	PublicKey       string
	Command         string
	HealthCheckAddr string
	HealthCheckPort int
	SSHServerAddr   string
	SSHServerPort   int
}

// Start rarukas-server
func Start(ctx context.Context, cfg *Config) error {

	// prepare for ssh server
	allowedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(cfg.PublicKey))
	if err != nil {
		return err
	}
	publicKeyOption := ssh.PublicKeyAuth(sshAuthHandler(allowedKey))
	ssh.Handle(sessionHandler(cfg.Command))

	// start
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error)

	// start health check
	hcAddr := fmt.Sprintf("%s:%d", cfg.HealthCheckAddr, cfg.HealthCheckPort)
	hcServer := &http.Server{
		Addr:    hcAddr,
		Handler: http.HandlerFunc(healthCheckHandler),
	}
	go func() {
		select {
		case errChan <- hcServer.ListenAndServe():
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := hcServer.Shutdown(shutdownCtx); err != nil {
				log.Println(err)
			}
		}
	}()

	// start ssh server
	sshAddr := fmt.Sprintf("%s:%d", cfg.SSHServerAddr, cfg.SSHServerPort)
	sshServer := &ssh.Server{Addr: sshAddr}
	sshServer.SetOption(publicKeyOption) // nolint return value not checked
	go func() {
		select {
		case errChan <- sshServer.ListenAndServe():
		case <-ctx.Done():
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sshServer.Shutdown(shutdownCtx); err != nil {
			log.Println(err)
		}
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ), // nolint return value not checked
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func sessionHandler(strCmd string) ssh.Handler {
	return func(s ssh.Session) {

		log.SetPrefix("[SSH]")
		log.SetFlags(log.LstdFlags)

		log.Printf("User %q connected from %q\n", s.User(), s.RemoteAddr().String())
		defer func() {
			log.Printf("User %q disconnected\n", s.User())
			log.SetPrefix("")
			log.SetFlags(0)
		}()
		args := []string{}
		if len(s.Command()) > 0 {
			args = []string{"-c", strings.Join(s.Command(), " ")}
		}

		cmd := exec.Command(strCmd, args...)

		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			f, err := pty.Start(cmd)
			if err != nil {
				panic(err)
			}
			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()

			go func() {
				// stdin
				io.Copy(f, s) // nolint return value not checked
			}()

			// stdout
			io.Copy(s, f) // nolint return value not checked
		} else {

			var err error
			var in io.WriteCloser
			var out io.ReadCloser
			var errOut io.ReadCloser

			// Prepare teardown function
			close := func() {
				in.Close() // nolint

				e1 := cmd.Wait()
				var exitStatus int32
				if e1 != nil {
					if e2, ok := e1.(*exec.ExitError); ok {
						if s, ok := e2.Sys().(syscall.WaitStatus); ok {
							exitStatus = int32(s.ExitStatus())
						} else {
							panic(errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus"))
						}
					}
				}
				var b bytes.Buffer
				binary.Write(&b, binary.BigEndian, exitStatus) // nolint
				s.SendRequest("exit-status", false, b.Bytes()) // nolint
				s.Close()                                      // nolint
			}

			in, err = cmd.StdinPipe()
			if err != nil {
				fmt.Fprint(s.Stderr(), err) // nolint
				close()
				return
			}

			out, err = cmd.StdoutPipe()
			if err != nil {
				fmt.Fprint(s.Stderr(), err) // nolint
				close()
				return
			}

			errOut, err = cmd.StderrPipe()
			if err != nil {
				fmt.Fprint(s.Stderr(), err) // nolint
				close()
				return
			}

			sigChan := make(chan ssh.Signal)
			s.Signals(sigChan)
			err = cmd.Start()
			if err != nil {
				fmt.Fprint(s.Stderr(), err) // nolint
				close()
				return
			}

			//pipe session to shell and visa-versa
			var once sync.Once
			go func() {
				io.Copy(s, out) // nolint
				once.Do(close)
			}()
			go func() {
				io.Copy(s.Stderr(), errOut) // nolint
				once.Do(close)
			}()
			go func() {
				io.Copy(in, s) // nolint
				once.Do(close)
			}()

			sig := <-sigChan
			log.Printf("session received signal[%s]", sig)
			once.Do(close)
		}
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK")) // nolint return value not checked
}

func sshAuthHandler(allowed ssh.PublicKey) ssh.PublicKeyHandler {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		if ctx.User() != "root" {
			return false
		}
		return ssh.KeysEqual(key, allowed)
	}
}
