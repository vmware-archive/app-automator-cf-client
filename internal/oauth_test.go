package internal_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    . "github.com/onsi/ginkgo/extensions/table"

    "github.com/pivotal-cf/eats-cf-client/internal/mocks"
    "github.com/pivotal-cf/eats-cf-client/internal"
    "github.com/pkg/errors"
    "net/http"
    "net/url"
)

var _ = Describe("Oauth", func() {
    type testContext struct {
        httpClient *mocks.HttpClient
    }

    var setup = func() (*internal.OauthClient, *testContext) {
        tc := &testContext{
            httpClient: mocks.NewHttpClient(),
        }
        client := internal.NewOauthClient(tc.httpClient, "https://example.com", "admin", "supersecret")

        return client, tc
    }

    It("gets a token", func() {
        client, tc := setup()

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
            client, tc := setup()
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
            tc.httpClient.Response = `im not json`
        }),
    )
})
