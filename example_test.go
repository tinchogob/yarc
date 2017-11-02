package yarc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/tinchogob/yarc/yaci"
	"github.com/tinchogob/yarc/yams"
)

func ExampleYarc() {
	ms, err := yams.New(8181)
	if err != nil {
		panic(err)
	}

	ms.Add(
		yams.Mock{
			Method:     http.MethodPost,
			URL:        "/items/1/ping/2?attributes=id",
			ReqBody:    []byte("{\"id\":\"ping\"}"),
			RespStatus: http.StatusOK,
			RespBody:   []byte("{\"id\":\"pong\"}"),
			Wait:       time.Millisecond * 100,
			Times:      2,
		},
	)

	yarc, err := New(
		Client(BaseClient(1, time.Second, time.Second)),
		Host("http://127.0.0.1:8181"),
		Path("/items/%s/ping/%s"),
		Header("Connection", "keep-alive"),
		Header("Cache-Control", "no-cache"),
		Trace(BaseTrace()),
		With(Debug(os.Stdout)),
		WithCache(yaci.New(time.Millisecond*100, 100)),
	)

	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)

	for i := 0; i < 1; i++ {
		time.Sleep(time.Millisecond * 200)
		wg.Add(1)
		go func() {

			//request body
			body := &struct {
				ID string `json:"id"`
			}{"ping"}

			//response struct to unmarshall response OK
			resp := &struct {
				ID string `json:"id"`
			}{}

			//response struct to unmarshall response NOK
			errBody := &struct {
				ID string `json:"id"`
			}{}

			res, err := yarc.Go(
				POST(),
				Header("X-Name", "Martin"),
				JSON(body),
				Params("1", "2"),
				Query("attributes", "id"),
				With(Context(context.Background())),
				ToJSON(resp, errBody),
			)

			wg.Done()

			if err != nil {
				fmt.Println(err.Error())
				if res != nil {
					fmt.Printf("%d - %s\n", res.StatusCode, errBody.ID)

				}
				return
			}

			fmt.Printf("%d - %s\n", res.StatusCode, resp.ID)
			//Output: 200 - pong

		}()
	}

	wg.Wait()
	ms.Close()
}
