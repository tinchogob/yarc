package yarc

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

func BaseClient(maxIdleConnsPerHost int, connectionTO time.Duration, requestTO time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout: connectionTO,
	}

	c := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   maxIdleConnsPerHost,
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ResponseHeaderTimeout: requestTO,
		},
	}

	return c
}

func BaseTrace() func(Options) (*httptrace.ClientTrace, error) {
	return func(opts Options) (*httptrace.ClientTrace, error) {
		return &httptrace.ClientTrace{
			GetConn: func(hostPort string) {
				fmt.Printf("%s: conn_request\n", opts.Path)

			},
			GotConn: func(connInfo httptrace.GotConnInfo) {
				if connInfo.Reused {
					fmt.Printf("%s: conn_got reused\n", opts.Path)
				} else {
					fmt.Printf("%s: conn_got not reused\n", opts.Path)
				}

			},
			ConnectDone: func(network, addr string, err error) {
				if err != nil {
					fmt.Printf("%s: conn_new fail\n", opts.Path)
				} else {
					fmt.Printf("%s: conn_new ok\n", opts.Path)
				}
			},
		}, nil
	}
}

func Debug(out io.Writer) withFunc {
	return func(opts Options, req *http.Request) *http.Request {
		if r, err := httputil.DumpRequest(req, true); err != nil {
			out.Write([]byte(err.Error()))
		} else {
			out.Write([]byte("<debug>\n"))
			out.Write(r)
			out.Write([]byte("\n</debug>\n"))
		}
		return req
	}
}

type histrixHTTPClient struct {
	path                  string
	timeout               int
	maxConcurrentRequests int
	errorPercentThreshold int
	client                *http.Client
}

func (hc histrixHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error
	var resp *http.Response

	hystrix.ConfigureCommand(hc.path, hystrix.CommandConfig{
		Timeout:               hc.timeout,
		MaxConcurrentRequests: hc.maxConcurrentRequests,
		ErrorPercentThreshold: hc.errorPercentThreshold,
	})

	hystrix.Do(hc.path, func() error {
		resp, err = hc.client.Do(req)
		return err
	}, nil)

	return resp, err
}

func Histrix(timeout int, maxConcurrentRequests int, errorPercentThreshold int, client *http.Client) optionFunc {
	return func(opts Options) (Options, error) {

		opts.Client = &http.Client{
			Transport: histrixHTTPClient{
				path:                  opts.Path,
				timeout:               timeout,
				maxConcurrentRequests: maxConcurrentRequests,
				errorPercentThreshold: errorPercentThreshold,
				client:                client,
			},
		}

		return opts, nil
	}
}
