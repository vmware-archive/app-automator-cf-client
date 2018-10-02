package internal

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"

    "github.com/pivotal-cf/eats-cf-client/models"
)

type capiRequestor interface {
    Do(method, path string, body string, v interface{}, opts ...models.HeaderOption) error
    GetPagedResources(path string, v Accumulator, opts ...models.HeaderOption) error
}

type CapiClient struct {
    requestor capiRequestor
}

func NewCapiClient(requestor capiRequestor) *CapiClient {
    return &CapiClient{
        requestor: requestor,
    }
}

func (c *CapiClient) Apps(query map[string]string) ([]models.App, error) {
    var apps []models.App
    err := c.requestor.GetPagedResources("/v3/apps?"+buildQuery(query), func(messages json.RawMessage) error {
        var page []models.App

        err := json.Unmarshal(messages, &page)
        if err != nil {
            return err
        }
        apps = append(apps, page...)
        return nil
    })
    return apps, err
}

func buildQuery(values map[string]string) string {
    query := url.Values{}
    for k, v := range values {
        query.Add(k, v)
    }
    return query.Encode()
}

func (c *CapiClient) Process(appGuid, processType string) (models.Process, error) {
    var p models.Process
    err := c.get(fmt.Sprintf("/v3/apps/%s/processes/%s", appGuid, processType), &p)
    return p, err
}

func (c *CapiClient) Scale(appGuid, processType string, instanceCount uint) error {
    path := fmt.Sprintf("/v3/apps/%s/processes/%s/actions/scale", appGuid, processType)
    body := fmt.Sprintf(`{"instances": %d}`, instanceCount)

    return c.requestor.Do(http.MethodPost, path, body, nil)
}

func (c *CapiClient) get(path string, v interface{}) error {
    return c.requestor.Do(http.MethodGet, path, "", v)
}

func (c *CapiClient) CreateTask(appGuid, command string, cfg models.TaskConfig, opts ...models.HeaderOption) (models.Task, error) {
    path := fmt.Sprintf("/v3/apps/%s/tasks", appGuid)

    taskRequest := struct {
        models.TaskConfig
        Command string `json:"command"`
    }{
        TaskConfig: cfg,
        Command:    command,
    }

    body, err := json.Marshal(&taskRequest)
    if err != nil {
        return models.Task{}, err
    }

    var task models.Task
    err = c.requestor.Do(http.MethodPost, path, string(body), &task, opts...)
    return task, err
}

func (c *CapiClient) Stop(appGuid string) error {
    path := fmt.Sprintf("/v3/apps/%s/actions/stop", appGuid)
    return c.requestor.Do(http.MethodPost, path, "", nil)
}
