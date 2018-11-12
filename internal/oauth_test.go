package internal_test

import (
    "errors"
    "net/http"
    "net/url"
    "time"

    "github.com/pivotal-cf/eats-cf-client/internal"
    "github.com/pivotal-cf/eats-cf-client/internal/mocks"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/ginkgo/extensions/table"
    . "github.com/onsi/gomega"
)

var _ = Describe("Oauth", func() {
    type testContext struct {
        httpClient *mocks.HttpClient
    }

    var setupHttpClient = func() *testContext {
        return &testContext{
            httpClient: mocks.NewHttpClient(),
        }
    }

    var setupUserClient = func() (*internal.OauthClient, *testContext) {
        tc := setupHttpClient()
        client := internal.NewUserOauthClient(tc.httpClient, "https://example.com", "admin", "supersecret")

        return client, tc
    }

    var setupClientCredsClient = func() (*internal.OauthClient, *testContext) {
        tc := setupHttpClient()
        client := internal.NewClientCredentialsOauthClient(tc.httpClient, "https://example.com", "client", "secret")

        return client, tc
    }

    Describe("Token()", func() {
        It("gets a token", func() {
            client, tc := setupUserClient()

            token, err := client.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("bearer lemons"))

            Expect(tc.httpClient.Reqs).To(Receive(Equal(mocks.HttpRequest{
                Body: url.Values{
                    "client_id":     {"cf"},
                    "client_secret": {""},
                    "username":      {"admin"},
                    "password":      {"supersecret"},
                    "grant_type":    {"password"},
                    "response_type": {"token"},
                }.Encode(),
                Url:    "https://example.com/oauth/token",
                Method: http.MethodPost,
                Headers: http.Header{
                    "Content-Type": {"application/x-www-form-urlencoded"},
                    "Accept":       {"application/json"},
                },
            })))
        })

        DescribeTable("errors",
            func(setupFunc func(*testContext)) {
                client, tc := setupUserClient()
                setupFunc(tc)

                _, err := client.Token()
                Expect(err).To(HaveOccurred())
            },
            Entry("httpClient errors", func(tc *testContext) {
                tc.httpClient.Err = errors.New("expected error")
            }),
            Entry("token request returns unexpected status", func(tc *testContext) {
                tc.httpClient.Status = http.StatusConflict
            }),
            Entry("oauth server returns invalid json", func(tc *testContext) {
                tc.httpClient.Responses <- "im not json"
            }),
        )
    })

    Describe("TokenWithExpiry()", func() {
        It("gets a token", func() {
            client, tc := setupUserClient()

            token, err := client.TokenWithExpiry()
            Expect(err).ToNot(HaveOccurred())
            Expect(token.Token).To(Equal("bearer lemons"))
            Expect(token.ExpiresAt).To(BeTemporally("~", time.Now().Add(24*time.Hour), time.Minute))

            Expect(tc.httpClient.Reqs).To(Receive(Equal(mocks.HttpRequest{
                Body: url.Values{
                    "client_id":     {"cf"},
                    "client_secret": {""},
                    "username":      {"admin"},
                    "password":      {"supersecret"},
                    "grant_type":    {"password"},
                    "response_type": {"token"},
                }.Encode(),
                Url:    "https://example.com/oauth/token",
                Method: http.MethodPost,
                Headers: http.Header{
                    "Content-Type": {"application/x-www-form-urlencoded"},
                    "Accept":       {"application/json"},
                },
            })))
        })

        DescribeTable("errors",
            func(setupFunc func(*testContext)) {
                client, tc := setupUserClient()
                setupFunc(tc)

                _, err := client.TokenWithExpiry()
                Expect(err).To(HaveOccurred())
            },
            Entry("httpClient errors", func(tc *testContext) {
                tc.httpClient.Err = errors.New("expected error")
            }),
            Entry("token request returns unexpected status", func(tc *testContext) {
                tc.httpClient.Status = http.StatusConflict
            }),
            Entry("oauth server returns invalid json", func(tc *testContext) {
                tc.httpClient.Responses <- "im not json"
            }),
        )
    })

    Describe("NewClientCredentialsOauthClient", func() {
        It("gets a token", func() {
            client, tc := setupClientCredsClient()

            token, err := client.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("bearer lemons"))

            Expect(tc.httpClient.Reqs).To(Receive(Equal(mocks.HttpRequest{
                Body: url.Values{
                    "client_id":     {"client"},
                    "client_secret": {"secret"},
                    "grant_type":    {"client_credentials"},
                    "response_type": {"token"},
                }.Encode(),
                Url:    "https://example.com/oauth/token",
                Method: http.MethodPost,
                Headers: http.Header{
                    "Content-Type": {"application/x-www-form-urlencoded"},
                    "Accept":       {"application/json"},
                },
            })))
        })
    })
})
