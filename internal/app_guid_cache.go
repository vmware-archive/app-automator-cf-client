package internal

import (
    "fmt"
    "sync"

    "github.com/pivotal-cf/eats-cf-client/models"
)

type appGetter func(query map[string]string) ([]models.App, error)

type AppGuidCache struct {
    get       appGetter
    spaceGuid string

    cache map[string]string
    mu    sync.RWMutex
}

func NewAppGuidCache(appGetter appGetter, spaceGuid string) *AppGuidCache {
    return &AppGuidCache{
        get:       appGetter,
        spaceGuid: spaceGuid,

        cache: make(map[string]string),
    }
}

func (c *AppGuidCache) Get(name string) (string, error) {
    c.mu.RLock()
    guid, ok := c.cache[name]
    c.mu.RUnlock()
    if ok {
        return guid, nil
    }

    err := c.refresh()
    if err != nil {
        return "", err
    }

    c.mu.RLock()
    guid, ok = c.cache[name]
    c.mu.RUnlock()
    if ok {
        return guid, nil
    }

    return "", fmt.Errorf("app '%s' not found", name)
}

func (c *AppGuidCache) refresh() error {
    apps, err := c.get(map[string]string{
        "space_guids": c.spaceGuid,
    })
    if err != nil {
        return err
    }

    newMap := make(map[string]string)
    for _, a := range apps {
        newMap[a.Name] = a.Guid
    }

    c.mu.Lock()
    c.cache = newMap
    c.mu.Unlock()

    return nil
}

func (c *AppGuidCache) Invalidate() {
    c.mu.Lock()
    c.cache = map[string]string{}
    c.mu.Unlock()
}

func (c *AppGuidCache) TryWithRefresh(appName string, f func(appGuid string) error) error {
    guid, err := c.Get(appName)
    if err != nil {
        return err
    }

    err = f(guid)
    if err != nil {
        return f(guid)
    }
    return nil
}
