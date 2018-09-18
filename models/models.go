package models

type App struct {
    Guid string `json:"guid"`
    Name string `json:"name"`
}

type Process struct {
    Instances int `json:"instances"`
}

type Task struct {
    Guid string `json:"guid"`
}

type TaskConfig struct {
    Name        string `json:"name,omitempty"`
    DiskInMB    uint   `json:"disk_in_mb,omitempty"`
    MemoryInMB  uint   `json:"memory_in_mb,omitempty"`
    DropletGUID string `json:"droplet_guid,omitempty"`
}
