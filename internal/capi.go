package internal

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"

    "github.com/pivotal-cf/eats-cf-client/models"
)

type capiDoer func(method, path string, body string, opts ...models.HeaderOption) ([]byte, error)

type CapiClient struct {
    Do capiDoer
}

func NewCapiClient(doer capiDoer) *CapiClient {
    return &CapiClient{
        Do: doer,
    }
}

func (c *CapiClient) Apps(query map[string]string) ([]models.App, error) {
    var appsResponse struct {
        Resources []models.App `json:"resources"`
    }
    err := c.get("/v3/apps?"+buildQuery(query), &appsResponse)
    return appsResponse.Resources, err
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

    _, err := c.Do(http.MethodPost, path, body)
    return err
}

func (c *CapiClient) get(path string, v interface{}) error {
    resp, err := c.Do(http.MethodGet, path, "")
    if err != nil {
        return err
    }

    return json.Unmarshal(resp, v)
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

    resp, err := c.Do(http.MethodPost, path, string(body), opts...)
    if err != nil {
        return models.Task{}, err
    }

    var task models.Task
    err = json.Unmarshal(resp, &task)

    return task, err
}

func (c *CapiClient) Stop(appGuid string) error {
    path := fmt.Sprintf("/v3/apps/%s/actions/stop", appGuid)
    _, err := c.Do(http.MethodPost, path, "")
    return err
}
