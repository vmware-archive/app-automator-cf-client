package internal_test

import (
    "errors"
    "net/http"

    "github.com/pivotal-cf/eats-cf-client/internal"
    "github.com/pivotal-cf/eats-cf-client/internal/mocks"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/ginkgo/extensions/table"
    . "github.com/onsi/gomega"
)

var _ = Describe("CapiDoer", func() {
    type testContext struct {
        httpClient  *mocks.HttpClient
        getTokenCalls int
        getTokenErr error
    }

    var setup = func() (*internal.CapiDoer, *testContext) {
        httpClient := mocks.NewHttpClient()
        httpClient.Response = `{"body": 1}`
        tc := &testContext{
            httpClient: httpClient,
        }
        client := internal.NewCapiDoer(tc.httpClient, "https://example.com", func() (string, error) {
            tc.getTokenCalls++
            return "bearer lemons", tc.getTokenErr
        })

        return client, tc
    }

    It("does the request", func() {
        client, tc := setup()

        err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil)
        Expect(err).ToNot(HaveOccurred())

        Expect(tc.httpClient.Reqs).To(Receive(Equal(mocks.HttpRequest{
            Body:   "I want lemons",
            Url:    "https://example.com/v2/lemons",
            Method: http.MethodGet,
            Headers: http.Header{
                "Authorization": {"bearer lemons"},
                "Content-Type":  {"application/json"},
            },
        })))
    })

    It("applies header options", func() {
        client, tc := setup()

        err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil, func(header *http.Header) {
            header.Add("Limes", "grapefruit")
        })
        Expect(err).ToNot(HaveOccurred())

        var req mocks.HttpRequest
        Expect(tc.httpClient.Reqs).To(Receive(&req))
        Expect(req.Headers).To(HaveKeyWithValue("Limes", []string{"grapefruit"}))
    })

    It("does not get auth token if provided in header options", func() {
        client, tc := setup()

        err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil, func(header *http.Header) {
            header.Add("Authorization", "grapefruit")
        })
        Expect(err).ToNot(HaveOccurred())

        var req mocks.HttpRequest
        Expect(tc.httpClient.Reqs).To(Receive(&req))
        Expect(tc.getTokenCalls).To(Equal(0))
        Expect(req.Headers).To(HaveKeyWithValue("Authorization", []string{"grapefruit"}))
    })

    It("does not return an error if body is nil", func() {
        client, tc := setup()
        tc.httpClient.Response = ""

        err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil)
        Expect(err).ToNot(HaveOccurred())
    })

    It("returns the body if interface is not nil", func() {
        client, _ := setup()

        resp := &struct{
            Body int `json:"body"`
        }{}
        err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", resp)
        Expect(err).ToNot(HaveOccurred())
        Expect(resp.Body).To(Equal(1))
    })

    It("returns an error if unmarshaling fails", func() {
        client, tc := setup()
        tc.httpClient.Response = `[invalid json}`

        resp := &struct{
            Body int `json:"body"`
        }{}
        err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", resp)
        Expect(err).To(HaveOccurred())
    })

    DescribeTable("errors",
        func(setupFunc func(*testContext)) {
            client, tc := setup()
            setupFunc(tc)

            err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil)
            Expect(err).To(HaveOccurred())
        },
        Entry("httpClient errors", func(tc *testContext) {
            tc.httpClient.Err = errors.New("expected error")
        }),
        Entry("request returns unexpected status", func(tc *testContext) {
            tc.httpClient.Status = http.StatusConflict
        }),
        Entry("get token returns an error", func(tc *testContext) {
            tc.getTokenErr = errors.New("expected error")
        }),
    )
})
