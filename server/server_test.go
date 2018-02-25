// +build !windows

package server

import (
	"context"
	"github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	allowPublicKey = []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDf9U6fuIvSuQVKgAmFbIdu37uxeaerCV5dWDudekJ/QO4bV+thPIAoUKIQP2gnz+rBoXwFuJLb6M4MXQ5Lr8t4eP5RRcrjz9NuZhZrwmAptzuw50dc7RfjDrHJ/Xn/VLTAUACOC1fEjcolR2ep9eCkPmXw0IBB0CHmWhIj+K+6Mg2H+JLLar1gTsnlO3hXQgeZ1LbYskISVBit3WZCXau2fVE8YyQf+lvmNGWukj3u1V3l/UiitDs0YB9K4Jg25lYxsBnR8C1tN/7HTZdVsj4vfXScCnjxlB7+JX5EmTpJGqnfB5TjVauJACDuxIPkFIGM0cgE99tdUd8v5xyKkc1h`)
	denyPublicKey  = []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDWCYbSBJevVAAEAISybvFMhRsTYxi0PUg36NzvNa/3CbS0m1J0SO8aGyDqP8gtpXm71ND4dzfrbqGnMVN80dT4Ho5/0pRSCkIBLbIReb4i4agRrQzRv9fa8RwzomFuiz3Ot4Iky/cr882E7LY7lCYysVgNPJRMGfyWGFWJ92EEpEdDAqoRmk+nnVBzBerQ6BQMoftVupA1m4w+heQtCVwL8y+tF2Es2h7H0dbc0ZfivEDJ9JopU6GD26FssBdSvDouS2hfupwJU9GTYcH9hYp2giR7FzrPsg4ydEh3qAG8A7REVF/mVWfEc99KFrHmxNlYGAWgW17nqKREQIMpEs/h`)
)

type testSSHContext struct {
	context.Context
	userName string
}

func (c *testSSHContext) User() string                    { return c.userName }
func (c *testSSHContext) SessionID() string               { return "" }
func (c *testSSHContext) ClientVersion() string           { return "" }
func (c *testSSHContext) ServerVersion() string           { return "" }
func (c *testSSHContext) RemoteAddr() net.Addr            { return nil }
func (c *testSSHContext) LocalAddr() net.Addr             { return nil }
func (c *testSSHContext) Permissions() *ssh.Permissions   { return nil }
func (c *testSSHContext) SetValue(key, value interface{}) {}

func TestHealthCheckServer(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	healthCheckHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []byte("OK"), w.Body.Bytes())
}

func TestSSHAuthHandler(t *testing.T) {

	// prepare ssh server
	allowedKey, _, _, _, err := ssh.ParseAuthorizedKey(allowPublicKey)
	if err != nil {
		t.Fatal(err)
	}
	deniedKey, _, _, _, err := ssh.ParseAuthorizedKey(denyPublicKey)
	if err != nil {
		t.Fatal(err)
	}

	handler := sshAuthHandler(allowedKey)

	t.Run("Invalid UserName", func(t *testing.T) {
		ctx := &testSSHContext{userName: "foobar"}
		assert.False(t, handler(ctx, allowedKey))
	})
	t.Run("Valid UserName", func(t *testing.T) {
		ctx := &testSSHContext{userName: "root"}
		assert.True(t, handler(ctx, allowedKey))
	})
	t.Run("Invalid key", func(t *testing.T) {
		ctx := &testSSHContext{userName: "foobar"}
		assert.False(t, handler(ctx, deniedKey))
	})
	t.Run("Invalid key with valid UserName", func(t *testing.T) {
		ctx := &testSSHContext{userName: "root"}
		assert.False(t, handler(ctx, deniedKey))
	})

}
