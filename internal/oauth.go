package internal

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
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

type TokenWithExpiry struct {
    Token     string
    ExpiresAt time.Time
}

func (c *OauthClient) Token() (string, error) {
    tokenResponse, err := c.TokenWithExpiry()
    if err != nil {
        return "", err
    }

    return tokenResponse.Token, nil
}

func (c *OauthClient) TokenWithExpiry() (TokenWithExpiry, error) {
    req, err := c.tokenRequest()
    if err != nil {
        return TokenWithExpiry{}, err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return TokenWithExpiry{}, err
    }

    if resp.StatusCode > 299 {
        err := fmt.Errorf("getting token returned unexpected status code %d", resp.StatusCode)
        return TokenWithExpiry{}, err
    }

    var tokenResponse struct {
        AccessToken string `json:"access_token"`
        TokenType   string `json:"token_type"`
        ExpiresIn   int    `json:"expires_in"`
    }
    err = decodeBody(resp, &tokenResponse)
    if err != nil {
        return TokenWithExpiry{}, err
    }

    return TokenWithExpiry{
        Token:     fmt.Sprintf("%s %s", tokenResponse.TokenType, tokenResponse.AccessToken),
        ExpiresAt: time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
    }, nil
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
