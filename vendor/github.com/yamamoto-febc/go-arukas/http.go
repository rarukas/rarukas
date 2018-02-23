package arukas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
)

// ErrorNotFound represents 404 not found error
type ErrorNotFound error

type httpAPI interface {
	get(path string) ([]byte, error)
	patch(path string, body interface{}) ([]byte, error)
	put(path string, body interface{}) ([]byte, error)
	post(path string, body interface{}) ([]byte, error)
	delete(path string) error
}

type httpClient struct {
	apiBaseURL *url.URL
	token      string
	secret     string
	userAgent  string
	trace      bool
	traceOut   io.Writer
	timeout    time.Duration
}

func (c *httpClient) get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

func (c *httpClient) patch(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPatch, path, body)
}

func (c *httpClient) post(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, body)
}

func (c *httpClient) put(path string, body interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPut, path, body)
}

func (c *httpClient) delete(path string) error {
	_, err := c.doRequest(http.MethodDelete, path, nil)
	return err
}

// newRequest Generates an realClient request for the Arukas API, but does not
// perform the request. The request's Accept header field will be
// set to:
//
//   Accept: application/vnd.api+json;
func (c *httpClient) newRequest(method, path string, body interface{}) (*http.Request, error) {
	var ctype string
	var rbody io.Reader

	switch t := body.(type) {
	case nil:
	case string:
		rbody = bytes.NewBufferString(t)
	case io.Reader:
		rbody = t
	case []byte:
		rbody = bytes.NewReader(t)
		ctype = "application/vnd.api+json"
	default:
		v := reflect.ValueOf(body)
		if !v.IsValid() {
			break
		}
		if v.Type().Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
			if !v.IsValid() {
				break
			}
		}

		j, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rbody = bytes.NewReader(j)
		ctype = "application/json"
	}
	requestURL := *c.apiBaseURL // shallow copy
	requestURL.Path += path
	if c.trace {
		fmt.Fprintf(c.traceOut, "Requesting: %s %s %s\n", method, requestURL.String(), rbody)
	}
	req, err := http.NewRequest(method, requestURL.String(), rbody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("User-Agent", c.userAgent)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	req.SetBasicAuth(c.token, c.secret)

	return req, nil
}

// do Sends a Arukas API request
func (c *httpClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var req *http.Request
	if body != nil {
		marshaled, err := json.Marshal(body)
		if err != nil {
			return []byte{}, err
		}

		if c.trace {
			fmt.Fprintln(c.traceOut, "json: ", string(marshaled))
		}
		body = marshaled
	}

	req, err := c.newRequest(method, path, body)
	if err != nil {
		return []byte{}, err
	}

	if c.trace {
		fmt.Fprintf(c.traceOut, "RequestHeader: %#v", req.Header)
	}

	return c.do(req)
}

// do Submits an realClient request
func (c *httpClient) do(req *http.Request) ([]byte, error) {

	httpClient := http.DefaultClient
	httpClient.Timeout = c.timeout

	res, err := httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close() // nolint

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}

	if c.trace {
		fmt.Fprintln(c.traceOut, "Status:", res.StatusCode)
		headers := make([]string, len(res.Header))
		for k := range res.Header {
			headers = append(headers, k)
		}
		sort.Strings(headers)
		for _, k := range headers {
			if k != "" {
				fmt.Fprintln(c.traceOut, k+":", strings.Join(res.Header[k], " "))
			}
		}
		fmt.Fprintln(c.traceOut, string(body))
	}

	if err = checkResponse(res); err != nil {
		return []byte{}, err
	}

	return body, err
}

// CheckResponse returns an error (of type *Error) if the response.
func checkResponse(res *http.Response) error {
	if res.StatusCode == 404 {
		return ErrorNotFound(fmt.Errorf("The resource does not found on the server: %s", res.Request.URL))
	} else if res.StatusCode >= 400 {
		return fmt.Errorf("Got realClient status code >= 400: %s", res.Status)
	}
	return nil
}
