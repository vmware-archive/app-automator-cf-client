package mocks

import (
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
    Err       error
    Status    int
    Responses chan string

    Reqs chan HttpRequest
}

func NewHttpClient() *HttpClient {
    return &HttpClient{
        Reqs:      make(chan HttpRequest, 100),
        Responses: make(chan string, 100),
        Status:    http.StatusOK,
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

    var resp string
    select {
    case resp = <-c.Responses:
    default:
        resp = `{"access_token": "lemons", "token_type": "bearer", "expires_in": 86400}`
    }

    respBody := ioutil.NopCloser(strings.NewReader(resp))
    return &http.Response{
        StatusCode: c.Status,
        Body:       respBody,
    }, c.Err
}
