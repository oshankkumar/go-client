package goclient

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/newrelic/go-agent"
)

type requestCanceler interface {
	CancelRequest(*http.Request)
}

func newDebugRoundTripper(next http.RoundTripper) http.RoundTripper {
	return &debugRoundTripper{next: next}
}

// debugRoundTripper is a middleware which can be use to log
// outgoing http request and the response from server.
type debugRoundTripper struct {
	next http.RoundTripper
}

func (d *debugRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if body, err := httputil.DumpRequestOut(req, true); err == nil {
		log.Println(string(body))
	}

	defer func(begin time.Time) {
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("took: %v", time.Since(begin))

		if body, err := httputil.DumpResponse(resp, true); err == nil {
			log.Println(string(body))
		}
	}(time.Now())

	return d.next.RoundTrip(req)
}

func (d *debugRoundTripper) CancelRequest(req *http.Request) {
	if canceler, ok := d.next.(requestCanceler); ok {
		canceler.CancelRequest(req)
	}
}

func newHystrixRoundTripper(cmdName string, next http.RoundTripper) http.RoundTripper {
	return &hystrixRoundTripper{cmdName: cmdName, next: next}
}

// hystrixRoundTripper is a middleware which executes any single HTTP
// transaction as a hystrix command.
//
// we can use hystrix.ConfigureCommand() to tweak the settings of
// hystrix command.
type hystrixRoundTripper struct {
	next    http.RoundTripper
	cmdName string
}

func (h *hystrixRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	err = hystrix.DoC(req.Context(), h.cmdName, func(ctx context.Context) error {
		resp, err = h.next.RoundTrip(req)
		return err
	}, nil)

	return
}

func (h *hystrixRoundTripper) CancelRequest(req *http.Request) {
	if canceler, ok := h.next.(requestCanceler); ok {
		canceler.CancelRequest(req)
	}
}

//// newRelicTransport creates an http.RoundTripper to instrument external requests.
func newRelicTransport(next http.RoundTripper) http.RoundTripper {
	return &newRelicRoundTripper{next: next}
}

type newRelicRoundTripper struct {
	next http.RoundTripper
}

func (n *newRelicRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	txn := newrelic.FromContext(req.Context())
	if txn == nil {
		return n.next.RoundTrip(req)
	}

	defer func() {
		if err != nil {
			txn.NoticeError(err)
		}
	}()
	return newrelic.NewRoundTripper(txn, n.next).RoundTrip(req)
}

func (n *newRelicRoundTripper) CancelRequest(req *http.Request) {
	if canceler, ok := n.next.(requestCanceler); ok {
		canceler.CancelRequest(req)
	}
}
