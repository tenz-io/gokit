package httpcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HttpMethod string

const (
	MethodGet    HttpMethod = http.MethodGet
	MethodPost   HttpMethod = http.MethodPost
	MethodPut    HttpMethod = http.MethodPut
	MethodDelete HttpMethod = http.MethodDelete
	MethodHead   HttpMethod = http.MethodHead
	MethodPatch  HttpMethod = http.MethodPatch
)

type (
	Params  map[string][]string
	Headers map[string]string
)

type Client interface {
	// JSON sends a POST request marshaling reqBody to JSON and unmarshalling respBody from JSON.
	// The request headers will be set to "application/json" in the Content-Type field.
	// respBody must be a pointer to a struct.
	JSON(ctx context.Context, url string, method HttpMethod, reqBody, respBody any, reqOpts ...RequestOption) (err error)
	// DoSimple sends an HTTP request and returns an HTTP response body as a byte slice.
	DoSimple(ctx context.Context, req *SimpleRequest) (respBody []byte, err error)
	// Do send an HTTP request and returns an HTTP response and error.
	// it delegates to http.Client.Do.
	Do(req *http.Request) (resp *http.Response, err error)
}

type SimpleRequest struct {
	Url     string
	Method  HttpMethod
	Headers Headers
	Params  Params
	ReqBody []byte
}

func (sq *SimpleRequest) validate() error {
	if sq.Url == "" {
		return fmt.Errorf("url is empty")
	}
	switch sq.Method {
	case MethodGet, MethodDelete, MethodHead:
	case MethodPost, MethodPut, MethodPatch:
	default:
		return fmt.Errorf("unsupported method: %s", sq.Method)
	}
	return nil
}

type RequestOption func(req *SimpleRequest)

func WithRequestParams(params Params) RequestOption {
	return func(req *SimpleRequest) {
		req.Params = params
	}
}

func WithRequestHeaders(headers Headers) RequestOption {
	return func(req *SimpleRequest) {
		req.Headers = headers
	}
}

func WithRequestBody(body []byte) RequestOption {
	return func(req *SimpleRequest) {
		req.ReqBody = body
	}
}

func NewSimpleRequest(url string, method HttpMethod, opts ...RequestOption) *SimpleRequest {
	req := &SimpleRequest{
		Url:    url,
		Method: method,
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

type client struct {
	cli *http.Client
}

func NewClient(cli *http.Client) Client {
	return &client{cli: cli}
}

func (c *client) JSON(ctx context.Context, url string, method HttpMethod, reqBody, respBody any, reqOpts ...RequestOption) error {
	var (
		err error
		req = NewSimpleRequest(url, method, reqOpts...)
	)

	// set head content type
	if req.Headers == nil {
		req.Headers = make(Headers)
	}
	req.Headers[HeaderNameContentType] = "application/json"

	req.ReqBody, err = json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshal request body: %w", err)
	}

	respData, err := c.DoSimple(ctx, req)
	if err != nil {
		return fmt.Errorf("error do simple request: %w", err)
	}

	if respData != nil && respBody != nil {
		err = json.Unmarshal(respData, respBody)
		if err != nil {
			return fmt.Errorf("error unmarshal response body: %w", err)
		}
	}

	return nil
}

func (c *client) DoSimple(ctx context.Context, req *SimpleRequest) (respBody []byte, err error) {
	var (
		reqBody io.Reader
	)

	if err = req.validate(); err != nil {
		return nil, fmt.Errorf("error validate request: %w", err)
	}

	if len(req.ReqBody) > 0 {
		reqBody = io.NopCloser(bytes.NewReader(req.ReqBody))
	}

	httpReq, err := c.newRequest(ctx, req.Method, req.Url, req.Params, req.Headers, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error create request: %w", err)
	}

	resp, err := c.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response status code: %d", resp.StatusCode)
	}

	respBody, err = c.readBody(resp)
	return

}

func (c *client) Do(req *http.Request) (resp *http.Response, err error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if c.cli == nil {
		return nil, fmt.Errorf("http client is nil")
	}
	return c.cli.Do(req)
}

func (c *client) newRequest(ctx context.Context,
	method HttpMethod,
	url string,
	params Params,
	headers Headers,
	body io.Reader,
) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(ctx, string(method), url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %w", method, err)
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, vars := range params {
			for _, v := range vars {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return req, nil
}

func (c *client) readBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}

	if resp.Body == nil {
		return nil, nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error read response body: %w", err)
	}

	return bs, nil
}
