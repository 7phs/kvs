package tst

import (
	"bytes"

	"github.com/valyala/fasthttp"
)

type Client interface {
	Add(key, value string) error
	Get(key string) (string, error)
}

type defaultClient struct {
	host   string
	client fasthttp.Client
}

func NewClient(host string) Client {
	return &defaultClient{
		host: host,
	}
}

func (o *defaultClient) Add(key, value string) error {
	_, err := o.do(fasthttp.MethodPost, "/"+key, []byte(value))

	return err
}

func (o *defaultClient) Get(key string) (string, error) {
	body, err := o.do(fasthttp.MethodGet, "/"+key)

	return string(body), err
}

func (o *defaultClient) do(method string, path string, body ...[]byte) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetRequestURI(o.host + path)
	req.Header.SetMethod(method)

	if len(body) > 0 {
		req.SetBody(body[0])
	}

	req.Header.Set("Accept-Encoding", "gzip")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the request
	err := o.client.Do(req, resp)
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode(); code != fasthttp.StatusOK {
		switch code {
		case fasthttp.StatusNotFound:
			return nil, ErrNotFound
		case fasthttp.StatusInsufficientStorage:
			return nil, ErrOutOfLimit
		}

		return nil, err
	}

	// Do we need to decompress the response?
	var respBody []byte

	if bytes.EqualFold(resp.Header.Peek("Content-Encoding"), []byte("gzip")) {
		respBody, err = resp.BodyGunzip()
		if err != nil {
			return nil, err
		}
	} else {
		respBody = resp.Body()
	}

	return respBody, nil
}
