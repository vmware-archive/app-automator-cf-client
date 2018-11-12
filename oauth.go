package client

import (
    "github.com/pivotal-cf/eats-cf-client/internal"
    "net/http"
)

type httpClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type OauthConfig struct {
    HttpClient httpClient
    OauthUrl   string

    Username string
    Password string

    Client       string
    ClientSecret string
}

func NewOauth(cfg OauthConfig) *internal.OauthClient {
    if cfg.Username != "" {
        return internal.NewUserOauthClient(
            cfg.HttpClient,
            cfg.OauthUrl,
            cfg.Username,
            cfg.Password,
        )
    }

    return internal.NewClientCredentialsOauthClient(
        cfg.HttpClient,
        cfg.OauthUrl,
        cfg.Client,
        cfg.ClientSecret,
    )
}
