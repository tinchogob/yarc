package yasci

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type stupid struct {
	cache map[string]value
	lock  *sync.RWMutex
	ttl   time.Duration
	size  int
}

type value struct {
	expiration time.Time
	status     int
	url        string
	body       []byte
}

// Yet Anothed (stupid) cache implementation
func New(ttl time.Duration, size int) *stupid {
	return &stupid{
		cache: make(map[string]value),
		lock:  new(sync.RWMutex),
		ttl:   ttl,
		size:  size,
	}
}

func (e *stupid) Get(key *http.Request) (*http.Response, error) {
	e.lock.RLock()
	v := e.cache[key.URL.String()]
	e.lock.RUnlock()

	if v.url == "" {
		return nil, nil
	}

	if v.expiration.Before(time.Now()) {
		e.lock.Lock()
		delete(e.cache, key.URL.String())
		e.size--
		e.lock.Unlock()
		return nil, nil
	}

	r := &http.Response{
		Status:     http.StatusText(v.status),
		StatusCode: v.status,
		Request:    key,
		Body:       ioutil.NopCloser(bytes.NewBuffer(v.body)),
	}

	return r, nil
}

func (e *stupid) Set(key *http.Request, response *http.Response) error {

	if !e.shouldSet(key, response) {
		return nil
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	response.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	v := value{
		status:     response.StatusCode,
		url:        key.URL.String(),
		body:       body,
		expiration: time.Now().Add(e.ttl),
	}

	e.lock.Lock()
	e.cache[key.URL.String()] = v
	e.size++
	e.lock.Unlock()

	return nil
}

func (e *stupid) shouldSet(key *http.Request, response *http.Response) bool {

	// If full, no cache
	if len(e.cache) >= e.size {
		return false
	}

	// if response status is not 2xx (success), no cache
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return false
	}

	return true
}
