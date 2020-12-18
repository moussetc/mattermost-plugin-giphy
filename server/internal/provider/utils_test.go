package provider

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type MockHttpClient struct {
	response            *http.Response
	testRequestFunc     func(*http.Request) bool
	lastRequestPassTest bool
}

func (c *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if c.testRequestFunc != nil {
		c.lastRequestPassTest = c.testRequestFunc(req)
	}
	return c.response, nil
}
func (c *MockHttpClient) Get(s string) (*http.Response, error) {
	return c.response, nil
}
func NewMockHttpClient(res *http.Response) *MockHttpClient {
	return &MockHttpClient{
		response:        res,
		testRequestFunc: nil,
	}
}

func newServerResponseOK(body string) *http.Response {
	r := &http.Response{
		StatusCode: 200,
	}
	if body != "" {
		r.Body = ioutil.NopCloser(bytes.NewBufferString(body))
	}
	return r
}

func newServerResponseKO(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
	}
}

func newServerResponseKOWithBody(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
	}
}
