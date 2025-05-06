package provider

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
)

type MockHTTPClient struct {
	response            *http.Response
	testRequestFunc     func(*http.Request) bool
	lastRequestPassTest bool
}

func (c *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if c.testRequestFunc != nil {
		c.lastRequestPassTest = c.testRequestFunc(req)
	}
	return c.response, nil
}

func (c *MockHTTPClient) Get(_ string) (*http.Response, error) {
	return c.response, nil
}

func NewMockHTTPClient(res *http.Response) *MockHTTPClient {
	return &MockHTTPClient{
		response:        res,
		testRequestFunc: nil,
	}
}

func newServerResponseOK(body string) *http.Response {
	r := &http.Response{
		StatusCode: 200,
	}
	if body != "" {
		r.Body = io.NopCloser(bytes.NewBufferString(body))
	}
	return r
}

func newServerResponseKO(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     strconv.Itoa(statusCode),
	}
}

func newServerResponseKOWithBody(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
