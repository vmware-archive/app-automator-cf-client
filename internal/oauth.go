package internal

import (
    "net/http"
    "encoding/json"
    "fmt"
    "net/url"
    "strings"
)

type httpClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type OauthClient struct {
    httpClient httpClient
    oauthUrl   string
    username   string
    password   string
}

func NewOauthClient(httpClient httpClient, oauthUrl, username, password string) *OauthClient {
    return &OauthClient{
        httpClient: httpClient,
        oauthUrl:   oauthUrl,
        username:   username,
        password:   password,
    }
}

func (c *OauthClient) Token() (string, error) {
    req, err := c.tokenRequest()
    if err != nil {
        return "", err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", err
    }

    if resp.StatusCode > 299 {
        return "", fmt.Errorf("getting token returned unexpected status code %d", resp.StatusCode)
    }

    var tokenResponse struct {
        AccessToken string `json:"access_token"`
        TokenType   string `json:"token_type"`
    }

    err = decodeBody(resp, &tokenResponse)
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("%s %s", tokenResponse.TokenType, tokenResponse.AccessToken), nil
}

func (c *OauthClient) tokenRequest() (*http.Request, error) {
    body := url.Values{
        "client_id":     {"cf"},
        "client_secret": {""},
        "username":      {c.username},
        "password":      {c.password},
        "grant_type":    {"password"},
        "response_type": {"token"},
    }.Encode()

    req, err := http.NewRequest(http.MethodPost, c.oauthUrl+"/oauth/token", strings.NewReader(body))
    if err != nil {
        return nil, err
    }
    req.Header.Add("Accept", "application/json")
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    return req, nil
}

func decodeBody(resp *http.Response, v interface{}) error {
    defer resp.Body.Close()
    return json.NewDecoder(resp.Body).Decode(v)
}
