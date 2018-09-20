package mocks

import (
    "io"
    "io/ioutil"
    "net/http"
    "strings"

    "github.com/onsi/gomega"
)

type HttpRequest struct {
    Body    string
    Url     string
    Method  string
    Headers http.Header
}

type HttpClient struct {
    Err      error
    Status   int
    Response string

    Reqs chan HttpRequest
}

func NewHttpClient() *HttpClient {
    return &HttpClient{
        Reqs:     make(chan HttpRequest, 100),
        Response: `{"access_token": "lemons", "token_type": "bearer", "expires_in": 86400}`,
        Status:   http.StatusOK,
    }
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
    var body []byte
    var err error
    if req.Body != nil {
        body, err = ioutil.ReadAll(req.Body)
        gomega.Expect(err).ToNot(gomega.HaveOccurred())
    }

    c.Reqs <- HttpRequest{
        Body:    string(body),
        Url:     req.URL.String(),
        Method:  req.Method,
        Headers: req.Header,
    }

    var respBody io.ReadCloser
    if c.Response != "" {
        respBody = ioutil.NopCloser(strings.NewReader(c.Response))
    }

    return &http.Response{
        StatusCode: c.Status,
        Body:       respBody,
    }, c.Err
}
