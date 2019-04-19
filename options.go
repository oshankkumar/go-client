package goclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

// OptFunc is a functor which receives a Client instance and override
// its members
type OptFunc func(client *Client) error

func WithHTTPClient(httpClient *http.Client) OptFunc {
	return func(client *Client) error {
		client.httpClient = httpClient
		return nil
	}
}

func WithTLSConfig(tlsOpts *TLSOptions) OptFunc {
	return func(client *Client) error {
		tlsConfig, err := NewTLSConfig(tlsOpts)
		if err != nil {
			return err
		}
		transport, ok := client.httpClient.Transport.(*http.Transport)
		if ok {
			transport.TLSClientConfig = tlsConfig
			return nil
		}
		return fmt.Errorf("failed to apply tls config on %T", client.httpClient.Transport)
	}
}

func WithHTTPTimeout(timeout time.Duration) OptFunc {
	return func(client *Client) error {
		client.httpClient.Timeout = timeout
		return nil
	}
}

func WithCustomHeaders(headers http.Header) OptFunc {
	return func(client *Client) error {
		client.headers = headers
		return nil
	}
}

func WithHystrix(conf *HystrixConfig) OptFunc {
	return func(client *Client) error {
		hystrix.ConfigureCommand(conf.Name, conf.CommandConfig)
		client.hystrixConfig = conf
		return nil
	}
}

func WithDebug() OptFunc {
	return func(client *Client) error {
		client.debug = true
		return nil
	}
}

func WithNewRelicEnable() OptFunc {
	return func(client *Client) error {
		client.enableNewRelic = true
		return nil
	}
}
