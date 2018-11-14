package internal_test

import (
    "errors"
    "sync"
    "time"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/pivotal-cf/eats-cf-client/internal"
)

var _ = Describe("Token cache", func() {
    Describe("Get()", func() {
        validToken := internal.TokenWithExpiry{
            Token:     "token",
            ExpiresAt: time.Now().Add(time.Hour),
        }

        It("gets token if not present", func() {
            var tokenRefreshed bool
            c := internal.NewTokenCache(
                func() (internal.TokenWithExpiry, error) {
                    tokenRefreshed = true
                    return validToken, nil
                },
            )

            token, err := c.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("token"))
            Expect(tokenRefreshed).To(BeTrue())
        })

        It("gets token from cache if present", func() {
            var tokenRefreshed int
            c := internal.NewTokenCache(
                func() (internal.TokenWithExpiry, error) {
                    tokenRefreshed++
                    return validToken, nil
                },
            )

            token, err := c.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("token"))

            token, err = c.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("token"))

            Expect(tokenRefreshed).To(Equal(1))
        })

        It("handles concurrent reads", func() {
            c := internal.NewTokenCache(
                func() (internal.TokenWithExpiry, error) {
                    return validToken, nil
                },
            )
            for i := 0; i < 50; i++ {
                go func() { c.Token() }()
            }
        })

        It("refreshes the token if it is close to expiring", func() {
            var tokenRefreshed int
            c := internal.NewTokenCache(
                func() (internal.TokenWithExpiry, error) {
                    tokenRefreshed++
                    return internal.TokenWithExpiry{
                        Token:     "token",
                        ExpiresAt: time.Now().Add(time.Minute),
                    }, nil
                },
            )

            token, err := c.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("token"))

            token, err = c.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("token"))

            Expect(tokenRefreshed).To(Equal(2))
        })

        It("refreshes the token only ONCE when it expires", func() {
            var tokenRefreshed int
            c := internal.NewTokenCache(
                func() (internal.TokenWithExpiry, error) {
                    tokenRefreshed++

                    time.Sleep(50 * time.Millisecond)
                    return internal.TokenWithExpiry{
                        Token:     "token",
                        ExpiresAt: time.Now().Add(time.Hour),
                    }, nil
                },
            )


            wg := &sync.WaitGroup{}
            wg.Add(1000)
            for i := 0; i < 1000; i++ {
                go func() {
                    defer GinkgoRecover()
                    defer wg.Done()

                    token, err := c.Token()
                    Expect(err).ToNot(HaveOccurred())
                    Expect(token).To(Equal("token"))
                }()
            }

            wg.Wait()

            Expect(tokenRefreshed).To(Equal(1))
        })

        It("returns an error if getting the token fails", func() {
            c := internal.NewTokenCache(
                func() (internal.TokenWithExpiry, error) {
                    return validToken, errors.New("expected")
                },
            )

            _, err := c.Token()
            Expect(err).To(HaveOccurred())
        })
    })
})
