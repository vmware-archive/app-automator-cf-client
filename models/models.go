package models

type App struct {
    Guid string `json:"guid"`
    Name string `json:"name"`
}

type Process struct {
    Instances int `json:"instances"`
}

type TaskConfig struct {
    Name        string `json:"name"`
    DiskInMB    uint   `json:"disk_in_mb"`
    MemoryInMB  uint   `json:"memory_in_mb"`
    DropletGUID string `json:"droplet_guid"`
}
