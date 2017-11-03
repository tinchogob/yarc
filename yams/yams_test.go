package yams

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestYams_Integration(t *testing.T) {
	cases := []struct {
		name      string
		port      int
		requests  func() []*http.Request
		mocks     func() []Mock
		responses func() []http.Response
	}{
		{
			name: "Error/NoMocks",
			port: 8181,
			requests: func() []*http.Request {
				r, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/", nil)
				return []*http.Request{r}
			},
			mocks: func() []Mock {
				return []Mock{}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: 404,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("no match for GET /"))),
					},
				}
			},
		},
		{
			name: "Error/NoMatchingMock",
			port: 8181,
			requests: func() []*http.Request {
				r, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", nil)
				return []*http.Request{r}
			},
			mocks: func() []Mock {
				return []Mock{
					Mock{
						Method: http.MethodGet,
						URL:    "/test_no_match",
					},
				}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: http.StatusNotFound,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("no match for GET /test"))),
					},
				}
			},
		},
		{
			name: "Error/NoMatchingHeadersMock",
			port: 8181,
			requests: func() []*http.Request {
				r, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", nil)
				return []*http.Request{r}
			},
			mocks: func() []Mock {
				return []Mock{
					Mock{
						Method:     http.MethodGet,
						URL:        "/test",
						ReqHeaders: http.Header(map[string][]string{"X-Tincho": []string{"is", "gonzalez"}}),
						RespStatus: http.StatusOK,
						RespBody:   []byte("tost"),
					},
				}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: http.StatusNotFound,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("no match for GET /test [headers mismatch]"))),
					},
				}
			},
		},
		{
			name: "Error/NoMatchingBodyMock",
			port: 8181,
			requests: func() []*http.Request {
				r, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", nil)
				return []*http.Request{r}
			},
			mocks: func() []Mock {
				return []Mock{
					Mock{
						Method:     http.MethodGet,
						URL:        "/test",
						ReqBody:    []byte("the request body to match with this mock"),
						RespStatus: http.StatusOK,
						RespBody:   []byte("tost"),
					},
				}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: http.StatusNotFound,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("no match for GET /test [body mismatch]"))),
					},
				}
			},
		},
		{
			name: "Error/NoMoreCallsExpected",
			port: 8181,
			requests: func() []*http.Request {
				r1, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", nil)
				r2, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", nil)
				return []*http.Request{r1, r2}
			},
			mocks: func() []Mock {
				return []Mock{
					Mock{
						Method:     http.MethodGet,
						URL:        "/test",
						RespStatus: http.StatusOK,
						RespBody:   []byte("tost"),
						Times:      1,
					},
				}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("tost"))),
					},
					{
						StatusCode: http.StatusBadRequest,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("no more calls were expected for (GET_/test)"))),
					},
				}
			},
		},
		{
			name: "OK/MatchingMock",
			port: 8181,
			requests: func() []*http.Request {
				r, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test/22/values?query=param", bytes.NewBuffer([]byte("the request body")))
				r.Header.Add("X-Yarc", "yams")
				r.Header.Add("X-Yarc", "rules")
				return []*http.Request{r}
			},
			mocks: func() []Mock {
				return []Mock{
					Mock{
						Method:     http.MethodGet,
						URL:        "/test/\\d+/values\\?query=\\w+$",
						ReqHeaders: http.Header(map[string][]string{"X-Yarc": []string{"yams", "rules"}}),
						ReqBody:    []byte("the request body"),
						RespStatus: http.StatusOK,
						RespBody:   []byte("tost"),
					},
				}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("tost"))),
					},
				}
			},
		},
		{
			name: "OK/MultipleMatchingMocks",
			port: 8181,
			requests: func() []*http.Request {
				r1, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", bytes.NewBuffer([]byte("ping")))
				r2, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/mambo", bytes.NewBuffer([]byte("ping_again")))
				return []*http.Request{r1, r2}
			},
			mocks: func() []Mock {
				return []Mock{
					Mock{
						Method:     http.MethodGet,
						URL:        "/test",
						ReqBody:    []byte("ping"),
						RespStatus: http.StatusOK,
						RespBody:   []byte("pong"),
					},
					Mock{
						Method:     http.MethodGet,
						URL:        "/mambo",
						ReqBody:    []byte("ping_again"),
						RespStatus: http.StatusOK,
						RespBody:   []byte("pong_again"),
					},
				}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("pong"))),
					},
					{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("pong_again"))),
					},
				}
			},
		},
		{
			name: "OK/MatchingMockMultipleTimes",
			port: 8181,
			requests: func() []*http.Request {
				r1, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", nil)
				r2, _ := http.NewRequest(http.MethodGet, "http://localhost:8181/test", nil)
				return []*http.Request{r1, r2}
			},
			mocks: func() []Mock {
				return []Mock{
					Mock{
						Method:     http.MethodGet,
						URL:        "/test",
						RespStatus: http.StatusOK,
						RespBody:   []byte("tost"),
						Times:      2,
					},
				}
			},
			responses: func() []http.Response {
				return []http.Response{
					{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("tost"))),
					},
					{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("tost"))),
					},
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server, err := New(c.port)
			defer server.Close()

			requests := c.requests()
			mocks := c.mocks()
			responses := c.responses()

			server.Add(mocks...)

			for i, r := range requests {
				client := &http.Client{}
				var res *http.Response
				res, err = client.Do(r)
				if err != nil {
					t.Fatal(err)
				}

				expectedResponse := responses[i]

				if res.StatusCode != expectedResponse.StatusCode {
					t.Errorf("%s: expected (%d) but got (%d)", c.name, expectedResponse.StatusCode, res.StatusCode)
				}

				defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}

				defer expectedResponse.Body.Close()
				expectedBody, err := ioutil.ReadAll(expectedResponse.Body)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(expectedBody, body) {
					t.Errorf("%s: expected (%s) but got (%s)", c.name, string(expectedBody), string(body))
				}

			}
		})
	}
}
