package yarc

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"strings"
)

type Yarc struct {
	opts Options
}

// Cache is Yarc's cache interface. Implementations must be goroutine safe.
// cache methods will be called for every request, so implementations
// should define when they want a hit and when they should set.
// There is an example implementation in yarc.StupidCache
type Cache interface {
	Get(key *http.Request) (*http.Response, error)
	Set(key *http.Request, response *http.Response) error
}

type nopCache struct{}

func (n nopCache) Get(key *http.Request) (*http.Response, error) {
	return nil, nil
}

func (n nopCache) Set(key *http.Request, response *http.Response) error {
	return nil
}

// Yarc builder. Option functions here will apply to
// every request made with this instance.
func New(optsFunc ...optionFunc) (*Yarc, error) {
	opts := Options{
		Headers: http.Header(make(map[string][]string)),
		cache:   nopCache{},
	}

	var err error
	for _, optFunc := range optsFunc {
		opts, err = optFunc(opts)
		if err != nil {
			return nil, err
		}
	}

	return &Yarc{
		opts: opts,
	}, nil
}

// Go sends an HTTP request and returns an HTTP response
// Option functions here will apply only to this request
func (y *Yarc) Go(optsFunc ...optionFunc) (*http.Response, error) {
	opts := y.opts
	var err error
	for _, optFunc := range optsFunc {
		opts, err = optFunc(opts)
		if err != nil {
			return nil, err
		}
	}

	url := getURL(opts)
	req, err := http.NewRequest(opts.Method, url, bytes.NewBuffer(opts.ReqBody))
	if err != nil {
		return nil, err
	}

	req.Host = opts.Host
	req.Header = opts.Headers

	for _, with := range opts.withs {
		req = with(opts, req)
	}

	if opts.trace != nil {
		t, err := opts.trace(opts)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), t))
	}

	var response *http.Response
	response, err = opts.cache.Get(req)
	if err != nil {
		return nil, err
	}

	if response == nil {

		response, err = opts.Client.Do(req)
		if err != nil {
			return nil, err
		}

		err = opts.cache.Set(req, response)
		if err != nil {
			return response, err
		}

	}

	if opts.resBody != nil {
		_, _, err = opts.resBody(response)
		if err != nil {
			return response, err
		}
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response, fmt.Errorf("error %d %s %s", response.StatusCode, opts.Method, url)
	}

	return response, nil
}

func getURL(opts Options) string {
	url := opts.Path
	for _, param := range opts.Params {
		url = strings.Replace(url, "%s", param, 1)
	}

	if len(opts.Query) > 0 {
		url += "?" + strings.Join(opts.Query, "&")
	}

	return opts.Host + url
}
