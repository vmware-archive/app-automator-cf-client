package client

type appsResponse struct {
    Resources []app `json:"resources"`
}

type app struct {
    Metadata struct {
        Guid string `json:"guid"`
    } `json:"metadata"`
    Entity struct {
        Name      string `json:"name"`
        Instances int    `json:"instances"`
    } `json:"entity"`
}
