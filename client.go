package goclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultClientTimeOut = time.Second * 5
)

// WrapRoundTripper wraps all middleware to be used with underlying
// http.RoundTripper.
func WrapRoundTripper(c *Client) http.RoundTripper {
	rt := c.httpClient.Transport

	if c.hystrixConfig != nil {
		rt = newHystrixRoundTripper(c.hystrixConfig.Name, rt)
	}
	if c.debug {
		rt = newDebugRoundTripper(rt)
	}
	if c.enableNewRelic {
		rt = newRelicTransport(rt)
	}
	return rt
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   defaultClientTimeOut,
		Transport: http.DefaultTransport,
	}
}

// NewWithOpts initializes a default REST Client.
// It takes optional functors to modify it while creating.
// For e.g., `NewClientWithOpts("<host-addr-of-abacus>",WithHTTPTimeout(...),WithTLSConfig(...))`
// We can also initialize custom http Client using WithHTTPClient(...)
// to send request.
func NewWithOpts(baseURL string, optFuncs ...OptFunc) (*Client, error) {
	client := &Client{
		baseURL:    baseURL,
		httpClient: defaultHTTPClient(),
		headers:    http.Header{},
	}

	for _, f := range optFuncs {
		if err := f(client); err != nil {
			return nil, err
		}
	}

	client.httpClient.Transport = WrapRoundTripper(client)
	return client, nil
}

// Client represents rest client
type Client struct {
	baseURL    string
	path       string
	method     string
	httpClient *http.Client
	headers    http.Header
	query      string

	hystrixConfig  *HystrixConfig
	debug          bool
	enableNewRelic bool
}

// Verb sets http Method to be use in calling server endpoint
func (c *Client) Verb(method string) *Client {
	c.method = method
	return c
}

// Path will set the endpoint to be invoke.
func (c *Client) Path(path string) *Client {
	c.path = path
	return c
}

func (c *Client) Query(q url.Values) *Client {
	if q != nil {
		c.query = q.Encode()
	}
	return c
}

func (c *Client) Headers(header http.Header) *Client {
	for key, val := range header {
		c.headers[key] = val
	}
	return c
}

func (c *Client) reset() {
	c.query = ""
	c.method = ""
	c.headers = http.Header{}
	c.path = ""
}

// Do sends an http Request to server and Decodes the response body in to successV if resp
// status code is 2XX or 3XX  otherwise it will decode it in failureV and return an error.
//
// An error is returned if caused by underlying http client.
//
// eg.. err := c.Verb(http.Method<>).Path("/path").Do(ctx, requestBody, responseBody,nil)
func (c *Client) Do(ctx context.Context, reqBody interface{}, successV interface{}, failureV interface{}) error {
	defer c.reset()
	body, err := encodeBody(reqBody)
	if err != nil {
		return err
	}

	req, err := c.buildRequest(c.method, c.path, c.query, c.headers, body)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if successV == nil {
			return nil
		}
		return json.NewDecoder(resp.Body).Decode(successV)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if failureV != nil {
		if err = json.Unmarshal(respBody, failureV); err != nil {
			return ApiError{resp.StatusCode, respBody, err.Error()}
		}
	}

	return ApiError{StatusCode: resp.StatusCode, RespBody: respBody}
}

func (c *Client) buildRequest(method string, path string, query string, headers http.Header, body io.Reader) (*http.Request, error) {
	urlPath, err := c.getAPIPath(path, query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, urlPath, body)
	if err != nil {
		return req, err
	}

	return c.addHeaders(req, headers), nil
}

func (c *Client) getAPIPath(path string, query string) (string, error) {
	if !strings.Contains(c.baseURL, "://") {
		c.baseURL = "http://" + c.baseURL
	}

	baseUri, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	return (&url.URL{Scheme: baseUri.Scheme, Host: baseUri.Host, Path: path, RawQuery: c.query}).String(), nil
}

func (c *Client) addHeaders(req *http.Request, headers http.Header) *http.Request {
	if headers != nil {
		for key, vals := range headers {
			req.Header[key] = vals
		}
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return req
}

func encodeBody(reqBody interface{}) (io.Reader, error) {
	if reqBody == nil {
		return nil, nil
	}

	body, err := json.Marshal(reqBody)
	return bytes.NewReader(body), err
}
