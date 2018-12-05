package internal_test

import (
    "encoding/json"
    "errors"
    "github.com/onsi/gomega/types"
    "net/http"

    "github.com/pivotal-cf/app-automator-cf-client/internal"
    "github.com/pivotal-cf/app-automator-cf-client/internal/mocks"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/ginkgo/extensions/table"
    . "github.com/onsi/gomega"
    . "github.com/onsi/gomega/gstruct"
)

var _ = Describe("CapiDoer", func() {
    type testContext struct {
        httpClient  *mocks.HttpClient
        getTokenCalls int
        getTokenErr error
    }

    var setup = func(respBodies ...string) (*internal.CapiDoer, *testContext) {
        httpClient := mocks.NewHttpClient()
        for _, resp := range respBodies {
            httpClient.Responses <- resp
        }
        tc := &testContext{
            httpClient: httpClient,
        }
        client := internal.NewCapiDoer(tc.httpClient, "https://example.com", func() (string, error) {
            tc.getTokenCalls++
            return "bearer lemons", tc.getTokenErr
        })

        return client, tc
    }

    Describe("Do()", func() {
        It("does the request", func() {
            client, tc := setup(`{"body": 1}`)

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
            client, tc := setup(`{"body": 1}`)

            err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil, func(header *http.Header) {
                header.Add("Limes", "grapefruit")
            })
            Expect(err).ToNot(HaveOccurred())

            var req mocks.HttpRequest
            Expect(tc.httpClient.Reqs).To(Receive(&req))
            Expect(req.Headers).To(HaveKeyWithValue("Limes", []string{"grapefruit"}))
        })

        It("does not get auth token if provided in header options", func() {
            client, tc := setup(`{"body": 1}`)

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
            client, _ := setup("")

            err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil)
            Expect(err).ToNot(HaveOccurred())
        })

        It("returns the body if interface is not nil", func() {
            client, _ := setup(`{"body": 1}`)

            resp := &struct {
                Body int `json:"body"`
            }{}
            err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", resp)
            Expect(err).ToNot(HaveOccurred())
            Expect(resp.Body).To(Equal(1))
        })

        It("returns an error if unmarshaling fails", func() {
            client, _ := setup(`[invalid json}`)

            resp := &struct {
                Body int `json:"body"`
            }{}
            err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", resp)
            Expect(err).To(HaveOccurred())
        })

        DescribeTable("errors",
            func(setupFunc func(*testContext), respCodeMatcher types.GomegaMatcher) {
                client, tc := setup(`{"body": 1}`)
                setupFunc(tc)

                err := client.Do(http.MethodGet, "/v2/lemons", "I want lemons", nil)
                Expect(err).To(HaveOccurred())
                Expect(err.ResponseCode).To(respCodeMatcher)
            },
            Entry("httpClient errors", func(tc *testContext) {
                tc.httpClient.Err = errors.New("expected error")
            }, BeZero()),
            Entry("request returns unexpected status", func(tc *testContext) {
                tc.httpClient.Status = http.StatusConflict
            }, Equal(http.StatusConflict)),
            Entry("get token returns an error", func(tc *testContext) {
                tc.getTokenErr = errors.New("expected error")
            }, BeZero()),
        )
    })

    Describe("Get()", func() {
        It("handles pagination", func() {
            client, tc := setup(firstPage, secondPage)

            type citrus struct {
                Citrus string `json:"citrus"`
            }

            var combinedResps []citrus
            err := client.GetPagedResources(
                "/v2/lemons",
                func(resources json.RawMessage) error {
                    var resp []citrus
                    Expect(json.Unmarshal(resources, &resp)).To(Succeed())
                    combinedResps = append(combinedResps, resp...)
                    return nil
                },
                func(header *http.Header) {
                    header.Add("Custom", "header")
                })
            Expect(err).ToNot(HaveOccurred())

            Expect(tc.httpClient.Reqs).To(Receive(MatchFields(IgnoreExtras, Fields{
                "Url":     Equal("https://example.com/v2/lemons"),
                "Headers": HaveKeyWithValue("Custom", []string{"header"}),
            })))

            Expect(tc.httpClient.Reqs).To(Receive(MatchFields(IgnoreExtras, Fields{
                "Url":     Equal("https://lemons.com/citrus"),
                "Headers": HaveKeyWithValue("Custom", []string{"header"}),
            })))

            Expect(combinedResps).To(ConsistOf(
                citrus{Citrus: "lemons"},
                citrus{Citrus: "limes"},
                citrus{Citrus: "grapefruit"},
            ))
        })

        It("returns an error if the visitor function does", func() {
            client, _ := setup(`{"body": 1}`)

            err := client.GetPagedResources(
                "/v2/lemons",
                func(resources json.RawMessage) error {
                    return errors.New("expected")
                },
            )
            Expect(err).To(HaveOccurred())
        })

        DescribeTable("errors",
            func(setupFunc func(*testContext)) {
                client, tc := setup(`{"body": 1}`)
                setupFunc(tc)

                err := client.GetPagedResources(
                    "/v2/lemons",
                    func(resources json.RawMessage) error {
                        return nil
                    },
                )
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
})

const (
    firstPage = `{
  "pagination": {
    "total_results": 3,
    "total_pages": 2,
    "next": {
      "href": "https://lemons.com/citrus"
    }
  },
  "resources": [
    { "citrus": "lemons" },
    { "citrus": "limes" }
  ]
}`

    secondPage = `{
  "pagination": {
    "total_results": 3,
    "total_pages": 2,
    "next": null
  },
  "resources": [
    { "citrus": "grapefruit" }
  ]
}`
)
