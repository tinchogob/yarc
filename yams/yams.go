package yams

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"time"
)

type server struct {
	ops        chan op
	mockServer *httptest.Server
}

type Mock struct {
	Method      string
	URL         string
	ReqHeaders  http.Header
	ReqBody     []byte
	RespStatus  int
	RespHeaders http.Header
	RespBody    []byte
	Wait        time.Duration
	Times       int
}

type op struct {
	code     string
	argument Mock
	response chan Mock
}

const add = "add"
const find = "find"
const flush = "flush"

// Yet Another Mockup Server
func New(port int) (*server, error) {
	s := &server{ops: make(chan op)}

	go s.runMocks()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	s.mockServer = httptest.NewUnstartedServer(s)
	s.mockServer.Listener.Close()
	s.mockServer.Listener = listener

	s.mockServer.Start()
	return s, nil
}

func (s *server) runMocks() {
	var mocks []Mock

	for operation := range s.ops {
		switch operation.code {
		case add:
			mocks = append(mocks, operation.argument)
		case flush:
			mocks = nil
		case find:
			for _, m := range mocks {
				if ok, err := regexp.MatchString(m.key(), operation.argument.key()); ok && err == nil {
					operation.response <- m
					m.Times--
				}
			}
		}
	}
}

func (s *server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	operation := op{find, Mock{Method: request.Method, URL: request.URL.String()}, make(chan Mock)}
	s.ops <- operation
	mock := <-operation.response

	err := mock.match(request)
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(err.Error()))
		return
	}

	if mock.Times <= 0 {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(fmt.Sprintf("no more calls were expected for (%s)", mock.key())))
		return
	}

	time.Sleep(mock.Wait)
	mock.write(response)
}

func (s *server) Add(mocks ...Mock) error {
	for _, m := range mocks {
		if m.Times == 0 {
			m.Times = 1
		}

		s.ops <- op{add, m, nil}
	}
	return nil
}

func (s *server) Close() {
	close(s.ops)
	s.mockServer.Close()
}

func (m *Mock) key() string {
	return fmt.Sprintf("%s_%s", m.Method, m.URL)
}

func (m *Mock) match(req *http.Request) error {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	if !bytes.Equal(m.ReqBody, body) {
		return fmt.Errorf("no match for %s %s [body mismatch]", req.Method, req.URL.String())
	}

	if len(m.ReqHeaders) > 0 && !reflect.DeepEqual(m.ReqHeaders, req.Header) {
		return fmt.Errorf("no match for %s %s [headers mismatch]", req.Method, req.URL.String())
	}

	return nil
}

func (m *Mock) write(response http.ResponseWriter) error {
	response.WriteHeader(m.RespStatus)
	for k, v := range m.RespHeaders {
		for _, vv := range v {
			response.Header().Add(k, vv)
		}
	}
	response.Write(m.RespBody)
	return nil
}
