package internal

import (
    "encoding/json"
    "net/http"
    "net/url"
    "fmt"
    "github.com/pivotal-cf/eats-cf-client/models"
)

type capiDoer func(method, path string, body string) ([]byte, error)

type CapiClient struct {
    Do capiDoer
}

func NewCapiClient(doer capiDoer) *CapiClient {
    return &CapiClient{
        Do: doer,
    }
}

func (c *CapiClient) Apps(query map[string]string) ([]models.App, error) {
    var appsResponse struct{ Resources []models.App `json:"resources"` }
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
    scalePath := fmt.Sprintf("/v3/apps/%s/processes/%s/actions/scale", appGuid, processType)
    scaleBody := fmt.Sprintf(`{"instances": %d}`, instanceCount)

    _, err := c.Do(http.MethodPost, scalePath, scaleBody)
    return err
}

func (c *CapiClient) get(path string, v interface{}) error {
    resp, err := c.Do(http.MethodGet, path, "")
    if err != nil {
        return err
    }

    return json.Unmarshal(resp, v)
}
