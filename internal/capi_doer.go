package internal

import (
    "encoding/json"
    "fmt"
    "github.com/pivotal-cf/eats-cf-client/models"
    "io/ioutil"
    "net/http"
    "strings"
)

type tokenGetter func() (string, error)

type CapiDoer struct {
    httpClient httpClient
    capiUrl    string
    getToken   tokenGetter
}

func NewCapiDoer(httpClient httpClient, capiUrl string, tokenGetter tokenGetter) *CapiDoer {
    return &CapiDoer{
        httpClient: httpClient,
        capiUrl:    capiUrl,
        getToken:   tokenGetter,
    }
}

func (c *CapiDoer) Do(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
    req, err := c.buildReq(method, path, body, opts...)
    if err != nil {
        return err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if code := resp.StatusCode; code > 299 || code < 200 {
        return fmt.Errorf("CAPI request (%s %s) returned unexpected status: %d", method, path, code) //TODO capi error
    }

    if v != nil {
        return json.NewDecoder(resp.Body).Decode(v)
    }

    return nil
}

func (c *CapiDoer) buildReq(method string, path string, body string, opts ...models.HeaderOption) (*http.Request, error) {
    req, err := http.NewRequest(method, c.capiUrl+path, ioutil.NopCloser(strings.NewReader(body)))
    if err != nil {
        return nil, err
    }

    req.Header.Add("Content-Type", "application/json")
    for _, o := range opts {
        o(&req.Header)
    }

    if _, ok := req.Header["Authorization"]; !ok {
        token, err := c.getToken()
        if err != nil {
            return nil, err
        }

        req.Header.Add("Authorization", token)
    }

    return req, err
}
