package yarc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
)

// Yarc options. You shouldn't use this directly but with option functions.
// Option functions will be applied in order.
// Each request may set its own option functions that will be applied after
// builder options and may override or add options.
type Options struct {
	Method  string
	Host    string
	Path    string
	Params  []string
	Query   []string
	ReqBody []byte
	Headers http.Header
	Client  *http.Client
	withs   []WithFunc
	resBody func(*http.Response) (interface{}, interface{}, error)
	trace   func(Options) (*httptrace.ClientTrace, error)
	cache   Cache
}

// Yarc Options modifier function. You should use this to
// access and change yarc options
type optionFunc func(opts Options) (Options, error)

// GET sets the request method to http.MethodGet.
func GET() optionFunc {
	return func(opts Options) (Options, error) {
		opts.Method = http.MethodGet
		return opts, nil
	}
}

// POST sets the request method to http.MethodPost.
func POST() optionFunc {
	return func(opts Options) (Options, error) {
		opts.Method = http.MethodPost
		return opts, nil
	}
}

// Host sets the request host+port to host.
func Host(host string) optionFunc {
	return func(opts Options) (Options, error) {
		opts.Host = host
		return opts, nil
	}
}

// Path sets the base path for this request.
// You should use a generic path with a format
// string (with %s) so that its generic and can
// identify all similar requests.
// Its intended to be used in conjunction with Params.
// For example:
// yarc.Go(Host("/ping/%s"),Params("me"))
// so yarc ends up calling "/ping/me".
func Path(path string) optionFunc {
	return func(opts Options) (Options, error) {
		opts.Path = path
		return opts, nil
	}
}

// Params sets the replace values for every %s
// in the request Path.
// Its intended to be used in conjunction with Path.
// For example:
// yarc.Go(Host("/ping/%s"),Params("me"))
// so yarc ends up calling "/ping/me".
func Params(params ...string) optionFunc {
	return func(opts Options) (Options, error) {
		opts.Params = params
		return opts, nil
	}
}

// Query adds key=value queryparam to the request.
// You should call Query as many time as params
// you have. They will be concatenated in the same
// order Query was called.
// TODO should be safe and sanitize this input
func Query(key string, value string) optionFunc {
	return func(opts Options) (Options, error) {
		opts.Query = append(opts.Query, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
		return opts, nil
	}
}

// Header adds name: value to the request headers
// You should call Header as many times as headers you
// want to set. Header will preserve previously setted
// headers.
func Header(key string, value string) optionFunc {
	return func(opts Options) (Options, error) {
		headers := http.Header(make(map[string][]string))
		for name, values := range opts.Headers {
			for _, value := range values {
				headers.Set(name, value)
			}
		}
		headers.Set(key, value)
		opts.Headers = headers
		return opts, nil
	}
}

// JSON sets the request body to the JSON marshall
// of the provided entity.
// It also adds "Content-Type: application/json"
// request header
func JSON(entity interface{}) optionFunc {
	return func(opts Options) (Options, error) {
		b, err := json.Marshal(entity)
		if err != nil {
			return opts, err
		}

		opts, err = Header("Content-Type", "application/json")(opts)
		if err != nil {
			return opts, err
		}

		return Body(b)(opts)
	}
}

// Body sets the request body to the provided []byte
// There a number of helper functions that simplifies
// usual cases such as marhsalling a json as the request body
// see JSON(), XML()
func Body(body []byte) optionFunc {
	return func(opts Options) (Options, error) {
		opts.ReqBody = body
		return opts, nil
	}
}

// ToJSON reads the response body and if OK
// tries to json.Uunmarshall it to body.
// If not OK tries to json.Uunmarshall it to errBody.
// It also adds "Accept: application/json" to the request
// headers.
func ToJSON(body interface{}, errBody interface{}) optionFunc {
	return func(opts Options) (Options, error) {
		opts, err := Header("Accept", "application/json")(opts)
		if err != nil {
			return opts, err
		}

		opts.resBody = func(response *http.Response) (interface{}, interface{}, error) {

			defer response.Body.Close()
			b, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, nil, err
			}

			var target interface{}
			if errBody != nil && response.StatusCode >= http.StatusBadRequest {
				target = errBody
			} else if body != nil && response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusBadRequest {
				target = body
			}

			if target == nil {
				return nil, nil, nil
			}

			err = json.Unmarshal(b, target)
			if err != nil {
				return nil, nil, fmt.Errorf("error unmarshalling response: %s\nresponse: %s\ntarget: %v", err.Error(), string(b), target)
			}

			return body, errBody, nil
		}

		return opts, nil
	}
}

// Client lets you use a custom http.Client.
// By default yarc will use the default http.Client.
func Client(client *http.Client) optionFunc {
	return func(opts Options) (Options, error) {
		opts.Client = client
		return opts, nil
	}
}

// Trace lets you build and use an httptrace.ClientTrace for this request.
// You should use this to gather metrics about each request.
// You have access to Options so you can group metrics by  Method, Host, Path, Etc.
func Trace(trace func(Options) (*httptrace.ClientTrace, error)) optionFunc {
	return func(opts Options) (Options, error) {
		opts.trace = trace
		return opts, nil
	}
}

// WithCache will make yarc to use an implementation of yarc.Cache.
// It MUST be gourutine safe.
func WithCache(c Cache) optionFunc {
	return func(opts Options) (Options, error) {
		opts.cache = c
		return opts, nil
	}
}

//With adds with to the request's with functions.
func With(with WithFunc) optionFunc {
	return func(opts Options) (Options, error) {
		q := len(opts.withs)
		withs := make([]WithFunc, q)
		for i, w := range opts.withs {
			withs[i] = w
		}
		withs = append(withs, with)
		opts.withs = withs
		return opts, nil
	}
}

// Provides complete access to the request.
// You can modify or even return a new request.
type WithFunc func(opts Options, req *http.Request) *http.Request

// Runs the request with ctx context.
func Context(ctx context.Context) WithFunc {
	return func(opts Options, req *http.Request) *http.Request {
		return req.WithContext(ctx)
	}
}

// Sets the request basic auth to username and password.
func BasicAuth(username string, password string) WithFunc {
	return func(opts Options, req *http.Request) *http.Request {
		req.SetBasicAuth(username, password)
		return req
	}
}
