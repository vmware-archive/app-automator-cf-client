package internal

import (
    "sync"
    "time"
)

type tokenWithExpiryGetter func() (TokenWithExpiry, error)

type TokenCache struct {
    get tokenWithExpiryGetter

    cachedToken TokenWithExpiry
    expiresAt   time.Time
    mu          sync.RWMutex
}

func NewTokenCache(tokenGetter tokenWithExpiryGetter) *TokenCache {
    return &TokenCache{
        get: tokenGetter,
    }
}

func (c *TokenCache) Token() (string, error) {
    c.mu.RLock()
    token := c.cachedToken
    c.mu.RUnlock()

    oneMinuteInFuture := time.Now().Add(time.Minute)
    if token.Token == "" || token.ExpiresAt.Before(oneMinuteInFuture) {
        return c.refresh()
    }

    return token.Token, nil
}

func (c *TokenCache) refresh() (string, error) {
    token, err := c.get()
    if err != nil {
        return "", err
    }

    c.mu.Lock()
    c.cachedToken = token
    c.mu.Unlock()

    return token.Token, nil
}
