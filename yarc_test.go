package yarc

import (
	"net/http"
	"testing"

	"github.com/tinchogob/yarc/yams"
)

func TestGo_basic(t *testing.T) {

	server, err := yams.New(8181)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.Add(yams.Mock{
		Method:     http.MethodGet,
		URL:        "/ping",
		RespStatus: http.StatusOK,
	})

	client, err := New(Host("http://localhost:8181"))
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.Go(Path("/ping"))
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("expected status 200")
	}
}

func TestGo_AdvancedOK(t *testing.T) {

	server, err := yams.New(8181)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.Add(yams.Mock{
		Method:     http.MethodPost,
		URL:        "/items/123\\?attributes=id&user=me",
		ReqBody:    []byte("{\"id\":\"123\"}"),
		RespStatus: http.StatusOK,
		RespBody:   []byte("{\"id\":\"123\"}"),
	})

	body := struct {
		ID string `json:"id"`
	}{"123"}

	client, err := New(
		Host("http://localhost:8181"),
		Path("/items/%s"),
	)
	if err != nil {
		t.Fatal(err)
	}

	item := struct {
		ID string `json:"id"`
	}{}

	response, err := client.Go(
		POST(),
		Params("123"),
		Query("attributes", "id"),
		Query("user", "me"),
		JSON(body),
		ToJSON(&item, nil),
	)

	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Error("expected status 200")
	}
}

func TestGo_AdvancedFail(t *testing.T) {

	server, err := yams.New(8181)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	server.Add(yams.Mock{
		Method:     http.MethodGet,
		URL:        "/items/123\\?attributes=id&user=me",
		RespStatus: http.StatusInternalServerError,
		RespBody:   []byte("{\"message\":\"error in the server\",\"cause\":\"la causa va aqui, claro que si\"}"),
	})

	client, err := New(
		Host("http://localhost:8181"),
		Path("/items/%s"),
	)
	if err != nil {
		t.Fatal(err)
	}

	respBody := struct {
		Message string `json:"message"`
		Cause   string `json:"cause"`
	}{}

	response, err := client.Go(
		GET(),
		Params("123"),
		Query("attributes", "id"),
		Query("user", "me"),
		ToJSON(nil, &respBody),
	)

	if err != nil && err.(*Yikes).Body != nil {
		if response.StatusCode != http.StatusInternalServerError {
			t.Error("expected status 500")
		}

		if err.Error() != "error 500 GET http://localhost:8181/items/123?attributes=id&user=me" {
			t.Errorf("expected (error 500 GET http://localhost:8181/items/123?attributes=id&user=me) but got (%s)", err.Error())
		}

		if respBody.Message != "error in the server" {
			t.Errorf("expected (error in the server) but got (%s)", respBody.Message)
		}
	} else if err != nil {
		t.Errorf("expected a server error")
	}
}
