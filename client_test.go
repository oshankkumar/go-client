package goclient

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_Do(t *testing.T) {
	testServer := httptest.NewServer(&testHandler{})
	client, err := NewWithOpts(
		testServer.URL,
		WithDebug(),
		WithNewRelicEnable(),
		WithHystrix(NewHystrixConfig("testcmd", hystrix.CommandConfig{})),
	)

	if err != nil {
		t.Fatalf("err: %s",err)
	}

	q := url.Values{"key1":[]string{"val1"},"key2":[]string{"val2"}}
	err = client.Verb(http.MethodGet).Path("/api/v1/testreq").Query(q).Do(context.Background(),nil,nil,nil)
	if err != nil {
		t.Fatalf("err: %s",err)
	}

}

type testHandler struct{}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() == "/api/v1/testreq?key1=val1&key2=val2" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}
