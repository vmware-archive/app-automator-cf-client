package client

import (
    "crypto/tls"
    "net/http"
    "strings"

    "github.com/pivotal-cf/eats-cf-client/internal"
    "github.com/pivotal-cf/eats-cf-client/models"
)

const (
    defaultTaskMemory = 10
    defaultTaskDisk   = 20

    defaultProcessType = "web"
)

type Oauth interface {
    Token() (string, error) //TODO just a func?
}

type Capi interface {
    Apps(query map[string]string) ([]models.App, error)
    Process(appGuid, processType string) (models.Process, error)
    Scale(appGuid, processType string, instanceCount uint) error
    CreateTask(appGuid, command string, cfg models.TaskConfig, opts ...models.HeaderOption) (models.Task, error)
    Stop(appGuid string) error
}

type AppGuidCache interface {
    TryWithRefresh(appName string, f func(appGuid string) error) error
}

type Client struct {
    CloudControllerUrl string
    SpaceGuid          string

    Oauth        Oauth
    Capi         Capi
    AppGuidCache AppGuidCache
}

type Config struct {
    CloudControllerUrl string
    SpaceGuid          string
    HttpClient         *http.Client
    Username           string
    Password           string
    TokenGetter        func() (string, error)
}

func Build() *Client {
    env := LoadEnv()
    httpClient := buildHttpClient(env)

    cfg := Config{
        CloudControllerUrl: env.CloudControllerApi,
        SpaceGuid:          env.VcapApplication.SpaceID,
        HttpClient:         httpClient,
    }

    credentials, ok := getUserProvidedCredentials(env.VcapServices.UserProvided)
    if ok {
        cfg.Username = credentials["username"]
        cfg.Password = credentials["password"]
    }

    return New(cfg)
}

func New(cfg Config) *Client {
    tokenEndpoint := strings.Replace(cfg.CloudControllerUrl, "api", "login", 1)
    oauth := NewTokenCache(OauthConfig{
        HttpClient: cfg.HttpClient,
        OauthUrl:   tokenEndpoint,
        Username:   cfg.Username,
        Password:   cfg.Password,
    })
    capi := internal.NewCapiClient(internal.NewCapiDoer(cfg.HttpClient, cfg.CloudControllerUrl, oauth.Token))

    return &Client{
        CloudControllerUrl: cfg.CloudControllerUrl,
        SpaceGuid:          cfg.SpaceGuid,
        Oauth:              oauth,
        Capi:               capi,
        AppGuidCache:       internal.NewAppGuidCache(capi.Apps, cfg.SpaceGuid),
    }
}

func buildHttpClient(env environment) *http.Client {
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
    return c.AppGuidCache.TryWithRefresh(appName, func(appGuid string) error {
        return c.Capi.Scale(appGuid, defaultProcessType, instanceTarget)
    })
}

func (c *Client) Process(appName, processType string) (models.Process, error) {
    var proc models.Process
    var err error
    err = c.AppGuidCache.TryWithRefresh(appName, func(appGuid string) error {
        proc, err = c.Capi.Process(appGuid, processType)
        return err
    })
    return proc, err
}

func (c *Client) CreateTask(appName, command string, cfg models.TaskConfig, opts ...models.HeaderOption) (models.Task, error) {
    if cfg.MemoryInMB == 0 {
        cfg.MemoryInMB = defaultTaskMemory
    }

    if cfg.DiskInMB == 0 {
        cfg.DiskInMB = defaultTaskDisk
    }
    if cfg.Name == "" {
        cfg.Name = command
    }

    var task models.Task
    var err error
    err = c.AppGuidCache.TryWithRefresh(appName, func(appGuid string) error {
        task, err = c.Capi.CreateTask(appGuid, command, cfg, opts...)
        return err
    })
    return task, err
}

func (c *Client) Stop(appName string) error {
    return c.AppGuidCache.TryWithRefresh(appName, func(appGuid string) error {
        return c.Capi.Stop(appGuid)
    })
}
