package internal_test

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/pivotal-cf/eats-cf-client/internal"
    "github.com/pivotal-cf/eats-cf-client/models"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Capi", func() {
    Describe("Apps()", func() {
        It("gets the apps", func() {
            mockDoer := newMockCapiGetter(func(path string, a internal.Accumulator, opts ...models.HeaderOption) error {
                Expect(path).To(And(
                    ContainSubstring("/v3/apps"),
                    ContainSubstring("lemons=limes"),
                    ContainSubstring("mangoes=limes"),
                ))

                Expect(a([]byte(appsPage1))).To(Succeed())
                Expect(a([]byte(appsPage2))).To(Succeed())
                return nil
            })
            c := internal.NewCapiClient(mockDoer)

            apps, err := c.Apps(map[string]string{
                "lemons":  "limes",
                "mangoes": "limes",
            })

            Expect(err).ToNot(HaveOccurred())
            Expect(apps).To(ConsistOf(
                models.App{Guid: "app-guid"},
                models.App{Guid: "app-guid-2"},
            ))
        })

        It("returns an error if requestor returns an error", func() {
            mockDoer := newMockCapiGetter(func(string, internal.Accumulator, ...models.HeaderOption) error {
                return errors.New("expected")
            })
            c := internal.NewCapiClient(mockDoer)

            _, err := c.Apps(nil)
            Expect(err).To(HaveOccurred())
        })
    })

    Describe("Process()", func() {
        It("gets the process", func() {
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                Expect(method).To(Equal(http.MethodGet))
                Expect(path).To(Equal("/v3/apps/app-guid/processes/process-type"))
                return json.Unmarshal([]byte(validProcessResponse), v)
            })
            c := internal.NewCapiClient(mockDoer)

            process, err := c.Process("app-guid", "process-type")
            Expect(err).ToNot(HaveOccurred())

            Expect(process).To(Equal(models.Process{
                Instances: 2,
            }))
        })

        It("returns an error if do returns an error", func() {
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                return errors.New("expected")
            })
            c := internal.NewCapiClient(mockDoer)

            _, err := c.Process("app-guid", "process-type")
            Expect(err).To(HaveOccurred())
        })
    })

    Describe("Scale()", func() {
        It("scales the process", func() {
            var called bool
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                called = true
                Expect(method).To(Equal(http.MethodPost))
                Expect(path).To(Equal("/v3/apps/app-guid/processes/process-type/actions/scale"))
                Expect(body).To(MatchJSON(`{ "instances": 5 }`))
                return nil
            })
            c := internal.NewCapiClient(mockDoer)

            Expect(c.Scale("app-guid", "process-type", 5)).To(Succeed())
            Expect(called).To(BeTrue())
        })

        It("returns an error if do returns an error", func() {
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                return errors.New("expected")
            })
            c := internal.NewCapiClient(mockDoer)

            Expect(c.Scale("app-guid", "process-type", 5)).ToNot(Succeed())
        })
    })

    Describe("CreateTask()", func() {
        It("creates a task", func() {
            var called bool
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                called = true
                Expect(method).To(Equal(http.MethodPost))
                Expect(path).To(Equal("/v3/apps/app-guid/tasks"))
                Expect(body).To(MatchJSON(`{
                    "command": "echo test",
                    "name": "lemons",
                    "disk_in_mb": 7,
                    "memory_in_mb": 30,
                    "droplet_guid": "droplet-guid"
                }`))
                return json.Unmarshal([]byte(validTaskResponse), v)
            })
            c := internal.NewCapiClient(mockDoer)

            task, err := c.CreateTask("app-guid", "echo test", models.TaskConfig{
                Name:        "lemons",
                DiskInMB:    7,
                MemoryInMB:  30,
                DropletGUID: "droplet-guid",
            })
            Expect(err).ToNot(HaveOccurred())
            Expect(called).To(BeTrue())
            Expect(task.Guid).To(Equal("task-guid"))
        })

        It("passes header options to doer", func() {
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                for _, o := range opts {
                    o(&http.Header{})
                }
                return json.Unmarshal([]byte(validTaskResponse), v)
            })
            c := internal.NewCapiClient(mockDoer)

            var headerOptionUsed bool
            opt := func(header *http.Header) {
                headerOptionUsed = true
            }
            _, err := c.CreateTask("app-guid", "echo test", models.TaskConfig{
                Name:        "lemons",
                DiskInMB:    7,
                MemoryInMB:  30,
                DropletGUID: "droplet-guid",
            }, opt)
            Expect(err).ToNot(HaveOccurred())

            Expect(headerOptionUsed).To(BeTrue())
        })

        It("returns an error if do returns an error", func() {
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                return errors.New("expected")
            })
            c := internal.NewCapiClient(mockDoer)

            _, err := c.CreateTask("app-guid", "command", models.TaskConfig{})
            Expect(err).To(HaveOccurred())
        })
    })

    Describe("Stop()", func() {
        It("stops the process", func() {
            var called bool
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                called = true
                Expect(method).To(Equal(http.MethodPost))
                Expect(path).To(Equal("/v3/apps/app-guid/actions/stop"))
                return nil
            })
            c := internal.NewCapiClient(mockDoer)

            Expect(c.Stop("app-guid")).To(Succeed())
            Expect(called).To(BeTrue())
        })

        It("returns an error if do returns an error", func() {
            mockDoer := newMockCapiDoer(func(method, path, body string, v interface{}, opts ...models.HeaderOption) error {
                return errors.New("expected")
            })
            c := internal.NewCapiClient(mockDoer)

            Expect(c.Stop("app-guid")).ToNot(Succeed())
        })
    })
})

const appsPage1 = `[ { "guid": "app-guid" } ]`
const appsPage2 = `[ { "guid": "app-guid-2" } ]`
const validProcessResponse = `{ "instances": 2 }`
const validTaskResponse = `{"guid": "task-guid"}`

type mockCapiRequestor struct {
    do  func(method string, path string, body string, v interface{}, opts ...models.HeaderOption) error
    get func(path string, v internal.Accumulator, opts ...models.HeaderOption) error
}

func newMockCapiDoer(do func(method, path string, body string, v interface{}, opts ...models.HeaderOption) error) *mockCapiRequestor {
    return &mockCapiRequestor{do: do}
}

func newMockCapiGetter(get func(path string, v internal.Accumulator, opts ...models.HeaderOption) error) *mockCapiRequestor {
    return &mockCapiRequestor{get: get}
}

func (d *mockCapiRequestor) Do(method, path string, body string, v interface{}, opts ...models.HeaderOption) error {
    return d.do(method, path, body, v, opts...)
}

func (d *mockCapiRequestor) GetPagedResources(path string, v internal.Accumulator, opts ...models.HeaderOption) error {
    return d.get(path, v, opts...)
}