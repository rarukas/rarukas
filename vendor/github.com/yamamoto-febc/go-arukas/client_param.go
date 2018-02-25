package arukas

import (
	"errors"
	"io"
	"time"

	"github.com/hashicorp/go-multierror"
)

// ClientParam represents parameters to use arukas.API
type ClientParam struct {
	APIBaseURL string
	Token      string
	Secret     string
	UserAgent  string
	Trace      bool
	TraceOut   io.Writer
	Timeout    time.Duration
}

func (p *ClientParam) validate() error {
	if p == nil {
		return errors.New("ClientParam is nil")
	}

	// check required param
	targets := map[string]string{
		"Toekn":  p.Token,
		"Secret": p.Secret,
	}

	var results error
	for k, v := range targets {
		err := validateRequired(k, v)
		if err != nil {
			results = multierror.Append(results, err)
		}
	}

	return results
}
