package arukas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"context"
)

const (
	defaultAPIBaseURL = "https://app.arukas.io/api"
	defaultUAFormat   = "go-arukas/v%s"
	defaultTimeout    = 30 * time.Second
)

// NewClient returns a new arukas API Client, requires an authorization key.
// You can generate a API key by visiting the Keys section of the Arukas
// control panel for your account.
func NewClient(p *ClientParam) (Client, error) {

	if err := p.validate(); err != nil {
		return nil, err
	}

	rawurl := p.APIBaseURL
	if rawurl == "" {
		rawurl = defaultAPIBaseURL
	}
	baseURL, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	userAgent := fmt.Sprintf(defaultUAFormat, Version)
	if p.UserAgent != "" {
		userAgent = p.UserAgent
	}

	var out io.Writer = os.Stdout
	if p.TraceOut != nil {
		out = p.TraceOut
	}
	if !p.Trace {
		out = ioutil.Discard
	}

	timeout := defaultTimeout
	if p.Timeout > 0 {
		timeout = p.Timeout
	}

	return &client{
		httpAPI: &httpClient{
			apiBaseURL: baseURL,
			token:      p.Token,
			secret:     p.Secret,
			userAgent:  userAgent,
			trace:      p.Trace,
			traceOut:   out,
			timeout:    timeout,
		},
	}, nil
}

// Client is Arukas API Client interface
type Client interface {
	ListApps() (*AppListData, error)
	ReadApp(id string) (*AppData, error)
	CreateApp(param *RequestParam) (*AppData, error)
	DeleteApp(id string) error

	ListServices() (*ServiceListData, error)
	ReadService(id string) (*ServiceData, error)
	UpdateService(id string, param *RequestParam) (*ServiceData, error)
	PowerOn(id string) error
	PowerOff(id string) error

	WaitForState(ctx context.Context, serviceID string, status string) error

	Version() string
}

// client implements arukas.api interface
type client struct {
	httpAPI httpAPI
}

// ListApp implements arukas.API interface
func (c *client) ListApps() (*AppListData, error) {
	data, err := c.httpAPI.get("/apps")
	if err != nil {
		return nil, err
	}

	var appListData AppListData
	err = json.Unmarshal(data, &appListData)
	if err != nil {
		return nil, err
	}

	return &appListData, nil
}

func (c *client) ReadApp(id string) (*AppData, error) {
	if err := validateID("ID", id); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/apps/%s", id)
	data, err := c.httpAPI.get(path)
	if err != nil {
		return nil, err
	}

	var appData AppData
	err = json.Unmarshal(data, &appData)
	if err != nil {
		return nil, err
	}

	return &appData, nil
}

// CreateApp implements arukas.API interface
func (c *client) CreateApp(param *RequestParam) (*AppData, error) {
	if param == nil {
		return nil, errors.New("param is nil")
	}
	err := param.ValidateForCreate()
	if err != nil {
		return nil, err
	}

	data, err := c.httpAPI.post("/apps", param.ToAppData())
	if err != nil {
		return nil, err
	}

	var appData AppData
	err = json.Unmarshal(data, &appData)
	if err != nil {
		return nil, err
	}

	return &appData, nil
}

// DeleteApp implements arukas.API interface
func (c *client) DeleteApp(id string) error {
	if err := validateID("ID", id); err != nil {
		return err
	}
	path := fmt.Sprintf("/apps/%s", id)
	return c.httpAPI.delete(path)
}

func (c *client) ListServices() (*ServiceListData, error) {
	data, err := c.httpAPI.get("/services")
	if err != nil {
		return nil, err
	}

	var serviceListData ServiceListData
	err = json.Unmarshal(data, &serviceListData)
	if err != nil {
		return nil, err
	}

	return &serviceListData, nil
}

func (c *client) ReadService(id string) (*ServiceData, error) {
	if err := validateID("ID", id); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/services/%s", id)
	data, err := c.httpAPI.get(path)
	if err != nil {
		return nil, err
	}

	var serviceData ServiceData
	err = json.Unmarshal(data, &serviceData)
	if err != nil {
		return nil, err
	}

	return &serviceData, nil
}

func (c *client) UpdateService(id string, param *RequestParam) (*ServiceData, error) {
	if err := validateID("ID", id); err != nil {
		return nil, err
	}
	if param == nil {
		return nil, errors.New("param is nil")
	}
	err := param.ValidateForUpdate()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/services/%s", id)
	data, err := c.httpAPI.patch(path, param.ToServiceData())
	if err != nil {
		return nil, err
	}

	var serviceData ServiceData
	err = json.Unmarshal(data, &serviceData)
	if err != nil {
		return nil, err
	}

	return &serviceData, nil
}

func (c *client) PowerOn(id string) error {
	if err := validateID("ID", id); err != nil {
		return err
	}
	path := fmt.Sprintf("/services/%s/power", id)
	_, err := c.httpAPI.post(path, nil)
	return err
}

func (c *client) PowerOff(id string) error {
	if err := validateID("ID", id); err != nil {
		return err
	}
	path := fmt.Sprintf("/services/%s/power", id)
	return c.httpAPI.delete(path)
}

func (c *client) WaitForState(ctx context.Context, serviceID string, status string) error {
	if err := validateID("ServiceID", serviceID); err != nil {
		return err
	}

	errChan := make(chan error, 1)

	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	go func() {
		for {
			s, err := c.ReadService(serviceID)
			if err != nil {
				errChan <- err
				return
			}
			if s.Status() == status {
				errChan <- nil
				return
			}
			<-ticker.C
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}

}

func (c *client) Version() string {
	return Version
}
