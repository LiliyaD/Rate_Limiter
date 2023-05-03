package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/LiliyaD/Rate_Limiter/config"
	"github.com/LiliyaD/Rate_Limiter/internal/static"
	"github.com/cornelk/hashmap/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestRequests(t *testing.T) {
	staticContentMock := newStaticContentMock()
	content, err := json.Marshal(staticContentMock)
	if err != nil {
		t.Error(err)
	}

	httpServer := NewHttpServer(&config.Config{
		SubnetPrefixLength: 24,
		TimeCooldownSec:    2,
		RateLimits:         config.RateLimits{RequestsCount: 2, TimeSec: 2},
	}, staticContentMock)

	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, httpServer.getStaticContent)
		if err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}
	}()

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return ln.Dial()
			},
		},
	}

	// Send first correct request
	req, err := http.NewRequest("GET", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, content, body)

	// Send request without header X-Forwarded-For
	req, err = http.NewRequest("GET", "http://test", nil)
	if err != nil {
		t.Error(err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)

	// Send second correct request
	req, err = http.NewRequest("GET", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.1")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, content, body)

	// Send third correct request, but error because of the limited number of requests for the subnet_1
	req, err = http.NewRequest("GET", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 429, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, []byte(fmt.Sprintf("Too many requests. Retry after %d seconds", httpServer.config.timeCooldownSec)), body)

	// Send fifth correct request for the subnet_2
	req, err = http.NewRequest("GET", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "222.45.67.1")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, content, body)

	// Send fourth correct request, but error because of the cooldown time for the subnet_1
	req, err = http.NewRequest("GET", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 429, resp.StatusCode)

	// Send first correct request after finished cooldown
	time.Sleep(2 * time.Second) // wait for cooldown
	req, err = http.NewRequest("GET", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, content, body)

	// Send wrong method request (POST)
	req, err = http.NewRequest("POST", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)

	// Send wrong method request (PUT)
	req, err = http.NewRequest("PUT", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)

	// Send wrong method request (DELETE)
	req, err = http.NewRequest("DELETE", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)

	// Send wrong method request (OPTIONS)
	req, err = http.NewRequest("OPTIONS", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)

	// Send wrong method request (HEAD)
	req, err = http.NewRequest("HEAD", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)

	// Send wrong method request ("TRACE")
	req, err = http.NewRequest("TRACE", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)

	// Send wrong method request ("CONNECT")
	req, err = http.NewRequest("CONNECT", "http://test", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("X-Forwarded-For", "123.45.67.89")

	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 400, resp.StatusCode)
}

type staticContentMock struct {
	Text string `json:"text"`
}

func newStaticContentMock() static.StaticContentI {
	return &staticContentMock{
		Text: "Test",
	}
}

func (h *staticContentMock) Get() ([]byte, error) {
	content, err := json.Marshal(h)
	if err != nil {
		err = errors.New("Response marshal error: " + err.Error())
	}
	return content, err
}
