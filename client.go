package client

import (
    "net/http"
    "github.com/pivotal-cf/eats-cf-client/internal"
    "strings"
    "crypto/tls"
    "github.com/pivotal-cf/eats-cf-client/models"
    "fmt"
)

type Oauth interface {
    Token() (string, error)
}

type Capi interface {
    Apps(query map[string]string) ([]models.App, error)
    Process(appGuid, processType string) (models.Process, error)
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

func (c *Client) Scale(appName string, instanceTarget uint) error {
    appGuid, err := c.appGuid(appName)
    if err != nil {
        return err
    }
    return c.Capi.Scale(appGuid, "web", instanceTarget)
}

func (c *Client) appGuid(appName string) (string, error) {
    apps, err := c.Capi.Apps(map[string]string{
        "names":       appName,
        "space_guids": c.SpaceGuid,
    })
    if err != nil {
        return "", err
    }

    if len(apps) != 1 {
        return "", fmt.Errorf("app %s not found in space %s", appName, c.SpaceGuid)
    }

    return apps[0].Guid, nil
}

func (c *Client) Process(appName, processType string) (models.Process, error) {
    appGuid, err := c.appGuid(appName)
    if err != nil {
        return models.Process{}, err
    }
    return c.Capi.Process(appGuid, processType)
}
