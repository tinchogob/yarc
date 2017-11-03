# yarc
Yet another rest client (for golang)

Yarc is a Go HTTP client library for creating and sending API requests.
Check [usage](#usage) or the [examples](examples) to learn how to yarc into your API client.

### Features

* Readable nice API
* Base/Request: Extend Yarc for different endpoints
* External cache support
* Native JSON support for sending/receiveng structs
* Clean access to http.Request && http.Response
* Collection of helpers for nice defaults
* Mock server for integration tests: [Yams](#https://github.com/tinchogob/yarc/tree/master/yams)
* ~~Stupid~~ Simple cache implementation: [Yasci](#https://github.com/tinchogob/yarc/tree/master/yasci)

## Install

    go get github.com/tinchogob/yarc

## Documentation

Read [GoDoc](https://godoc.org/github.com/tinchogob/yarc)

## Usage
Yarc uses dave's cheney [funcional options pattern](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
There a number of functional option functions to customize each client/request.

The can be applied to Yarc's constructor or to Yarc's request executor.

For example, create a base Yarc client with a host and then extend this client to do a request for method+path.

```go
client, err := Yarc.New(Host("https://api.mercadolibre.com))
if err != nil {
  panic(err)
}

r, err := client.Go(
  GET(),
  Path("/items/1234567"),
)
```

### Query, Headers and on 

```go
r, err := client.Go(
  GET(),
  Header("Connection", "keep-alive"),
  Path("/items/1234567"),
  Query("attributes","id,permalink"),
)
```

### Body

While dealing with JSON requests/response APIS, Yarcs provides some nice helpers out of the box.

#### Posting JSON

Define [JSON tagged structs](https://golang.org/pkg/encoding/json/). Use `JSON` to JSON encode a struct as the Body on requests.

```go
type ItemRequest struct {
    ID     string   `json:"id"`
    Permalink  string   `json:"permalink"`
}

item := ItemRequest{
  ID: "123",
  Permalink: "http://permalink.com",
}

r, err := client.Go(
  POST(),
  Header("Connection", "keep-alive"),
  Path("/items/1234567"),
  Query("attributes","id,permalink"),
  JSON(item),
)

```

Requests will include an `application/json` Content-Type header.

#### ToJSON

Define JSON structs to decode responses. Use `ToJSON(body interface{}, errBody interface{})` to decode the response body into `body` if 2xx or else to `errBody`.

```go
type Item struct {
    ID     string   `json:"id"`
    Permalink  string   `json:"permalink"`
}

type ApiError struct {
    Message string  `json:"message"`
    Cause string    `json:"cause"`
}

item := Item{}
errB := ApiError{}

r, err := client.Go(
  GET(),
  Header("Connection", "keep-alive"),
  Path("/items/1234567"),
  Query("attributes","id,permalink"),
  ToJSON(&item, &errB),
)

fmt.Println(item, errB)
```

### Accesing to http.Request

Yarcs provides an extension point called `With` to change/enhance each request

```go
r, err := client.Go(
  GET(),
  Path("/items/1234567"),
  With(Context(context.Background())),
)
```

## License

[MIT License](LICENSE)
