package client_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    . "github.com/onsi/gomega/gstruct"
    "github.com/pivotal-cf/eats-cf-client/internal/mocks"
    "net/url"

    . "github.com/pivotal-cf/eats-cf-client"
)

var _ = Describe("Oauth", func() {
    type testContext struct {
        httpClient *mocks.HttpClient
    }

    var setup = func() *testContext {
        return &testContext{
            httpClient: mocks.NewHttpClient(),
        }
    }

    Describe("NewOauth()", func() {
        It("uses user credentials if set", func() {
            tc := setup()

            client := NewOauth(OauthConfig{
                HttpClient: tc.httpClient,
                OauthUrl:   "https://example.com",
                Username:   "admin",
                Password:   "supersecret",
            })
            token, err := client.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("bearer lemons"))

            Expect(tc.httpClient.Reqs).To(Receive(MatchFields(IgnoreExtras,
                Fields{
                    "Body": Equal(url.Values{
                        "client_id":     {"cf"},
                        "client_secret": {""},
                        "username":      {"admin"},
                        "password":      {"supersecret"},
                        "grant_type":    {"password"},
                        "response_type": {"token"},
                    }.Encode()),
                },
            )))
        })

        It("uses client credentials if set", func() {
            tc := setup()

            client := NewOauth(OauthConfig{
                HttpClient:   tc.httpClient,
                OauthUrl:     "https://example.com",
                Client:       "client",
                ClientSecret: "secret",
            })
            token, err := client.Token()
            Expect(err).ToNot(HaveOccurred())
            Expect(token).To(Equal("bearer lemons"))

            Expect(tc.httpClient.Reqs).To(Receive(MatchFields(IgnoreExtras,
                Fields{
                    "Body": Equal(url.Values{
                        "client_id":     {"client"},
                        "client_secret": {"secret"},
                        "grant_type":    {"client_credentials"},
                        "response_type": {"token"},
                    }.Encode()),
                },
            )))
        })
    })
})
