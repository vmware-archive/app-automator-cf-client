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

    sync.Mutex
}

func NewTokenCache(tokenGetter tokenWithExpiryGetter) *TokenCache {
    return &TokenCache{
        get: tokenGetter,
    }
}

func (c *TokenCache) Token() (string, error) {
    oneMinuteInFuture := time.Now().Add(time.Minute)

    c.Lock()
    defer c.Unlock()

    token := c.cachedToken
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

    c.cachedToken = token

    return token.Token, nil
}
