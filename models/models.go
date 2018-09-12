package models

type App struct {
    Guid string `json:"guid"`
}

type Process struct {
    Instances int `json:"instances"`
}
