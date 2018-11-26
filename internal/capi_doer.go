package internal

import (
    "encoding/json"
    "fmt"
    "github.com/pivotal-cf/app-automator-cf-client/models"
    "io"
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
    return c.doUrl(method, c.capiUrl+path, body, v, opts...)
}

func (c *CapiDoer) doUrl(method, url, body string, v interface{}, opts ...models.HeaderOption) error {
    req, err := c.buildReq(method, url, body, opts...)
    if err != nil {
        return err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if code := resp.StatusCode; code > 299 || code < 200 {
        return fmt.Errorf("CAPI request (%s %s) returned unexpected status (%d): %s",
            method, url, code, decodeCapiErr(resp.Body))
    }

    if v != nil {
        return json.NewDecoder(resp.Body).Decode(v)
    }

    return nil
}

func decodeCapiErr(body io.Reader) error {
    var capiErr struct {
        Title  string `json:"title"`
        Detail string `json:"detail"`
    }
    err := json.NewDecoder(body).Decode(&capiErr)
    if err != nil {
        return fmt.Errorf("cannot decode CAPI error")
    }

    return fmt.Errorf("%s (%s)", capiErr.Title, capiErr.Detail)
}

func (c *CapiDoer) buildReq(method string, url string, body string, opts ...models.HeaderOption) (*http.Request, error) {
    req, err := http.NewRequest(method, url, ioutil.NopCloser(strings.NewReader(body)))
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

type paginatedResp struct {
    Resources  json.RawMessage `json:"resources"`
    Pagination pagination      `json:"pagination"`
}

type pagination struct {
    Next *struct {
        Href string `json:"href"`
    } `json:"next"`
}

type Accumulator func(json.RawMessage) error

func (c *CapiDoer) GetPagedResources(path string, a Accumulator, opts ...models.HeaderOption) error {
    var err error
    url := c.capiUrl + path
    for url != "" {
        url, err = c.getPage(url, a, opts...)
        if err != nil {
            return err
        }
    }

    return nil
}

func (c *CapiDoer) getPage(url string, a Accumulator, opts ...models.HeaderOption) (string, error) {
    var page = &paginatedResp{}
    err := c.doUrl(http.MethodGet, url, "", page, opts...)
    if err != nil {
        return "", err
    }

    err = a(page.Resources)
    if err != nil {
        return "", err
    }

    if page.Pagination.Next != nil {
        return page.Pagination.Next.Href, nil
    }

    return "", nil
}
