package client

import (
    "net/http"
    "github.com/pivotal-cf/eats-cf-client/internal"
    "strings"
    "fmt"
    "crypto/tls"
)

type Oauth interface {
    Token() (string, error)
}

type Capi interface {
    Apps(query map[string]string) ([]internal.App, error)
    Process(appGuid, processType string) (internal.Process, error)
    Scale(appGuid, processType string, instanceCount uint) error
}

type Client struct {
    CloudControllerUrl string
    SpaceGuid          string

    Oauth Oauth
    Capi  Capi
}

func New(env Environment, username, password string) *Client {
    httpClient := buildHttpClient(env)

    tokenEndpoint := strings.Replace(env.CloudControllerApi, "api", "login", 1)
    oauth := internal.NewOauthClient(httpClient, tokenEndpoint, username, password)

    return &Client{
        CloudControllerUrl: env.CloudControllerApi,
        SpaceGuid:          env.VcapApplication.SpaceID,

        Oauth: oauth,
        Capi:  internal.NewCapiClient(internal.NewCapiDoer(httpClient, env.CloudControllerApi, oauth.Token).Do),
    }
}

func buildHttpClient(env Environment) *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: env.SkipSslValidation,
            },
        },
        Timeout: env.HttpTimeout,
    }
}

func (c *Client) Scale(appName string, instanceDelta int) error {
    apps, err := c.Capi.Apps(map[string]string{
        "names":       appName,
        "space_guids": c.SpaceGuid,
    })
    if err != nil {
        return err
    }

    if len(apps) != 1 {
        return fmt.Errorf("app %s not found in space %s", appName, c.SpaceGuid)
    }

    process, err := c.Capi.Process(apps[0].Guid, "web")
    if err != nil {
        return err
    }

    //TODO shouldn't be negative
    t := process.Instances + instanceDelta

    return c.Capi.Scale(apps[0].Guid, "web", uint(t))
}
