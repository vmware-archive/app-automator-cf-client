package client_test

import (
    "github.com/pivotal-cf/eats-cf-client"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/ginkgo/extensions/table"
    . "github.com/onsi/gomega"
    "github.com/pkg/errors"
    "github.com/pivotal-cf/eats-cf-client/models"
)

var _ = Describe("Client", func() {
    Describe("Scale()", func() {
        DescribeTable("errors", func(modifyCapi func(*mockCapi)) {
            capi := &mockCapi{
                apps:       []models.App{{Guid: "1"}},
            }
            modifyCapi(capi)

            c := client.Client{
                Oauth: &mockOauth{},
                Capi:  capi,
            }

            Expect(c.Scale("lemons", 2)).ToNot(Succeed())
        },
            Entry("get apps returns an error", func(capi *mockCapi) {
                capi.appsErr = errors.New("expected error")
            }),
            Entry("get apps returns no apps", func(capi *mockCapi) {
                capi.apps = []models.App{}
            }),
            Entry("get apps returns multiple apps", func(capi *mockCapi) {
                capi.apps = []models.App{
                    {Guid: "1"},
                    {Guid: "2"},
                }
            }),
            Entry("scale returns an error", func(capi *mockCapi) {
                capi.scaleErr = errors.New("expected")
            }),
        )
    })

    Describe("Process()", func() {
        DescribeTable("errors", func(modifyCapi func(*mockCapi)) {
            capi := &mockCapi{
                apps:       []models.App{{Guid: "1"}},
                process:    models.Process{Instances: 2},
            }
            modifyCapi(capi)

            c := client.Client{
                Oauth: &mockOauth{},
                Capi:  capi,
            }

            _, err := c.Process("lemons", "web")
            Expect(err).To(HaveOccurred())
        },
            Entry("get apps returns an error", func(capi *mockCapi) {
                capi.appsErr = errors.New("expected error")
            }),
            Entry("get apps returns no apps", func(capi *mockCapi) {
                capi.apps = []models.App{}
            }),
            Entry("get apps returns multiple apps", func(capi *mockCapi) {
                capi.apps = []models.App{
                    {Guid: "1"},
                    {Guid: "2"},
                }
            }),
            Entry("process returns an error", func(capi *mockCapi) {
                capi.processErr = errors.New("expected")
            }),
        )
    })
})

type mockOauth struct {
    err error
}

func (o *mockOauth) Token() (string, error) {
    return "bearer token", o.err
}

type mockCapi struct {
    apps    []models.App
    appsErr error

    process    models.Process
    processErr error

    scaleErr error
}

func (c *mockCapi) Apps(query map[string]string) ([]models.App, error) {
    return c.apps, c.appsErr
}

func (c *mockCapi) Process(appGuid, processType string) (models.Process, error) {
    return c.process, c.processErr
}

func (c *mockCapi) Scale(appGuid, processType string, instanceCount uint) error {
    return c.scaleErr
}
