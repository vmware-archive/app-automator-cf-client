package internal_test

import (
    "errors"
    "net/http"

    "github.com/pivotal-cf/eats-cf-client/internal"
    "github.com/pivotal-cf/eats-cf-client/models"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/ginkgo/extensions/table"
    . "github.com/onsi/gomega"
)

var _ = Describe("Capi", func() {
    Describe("Apps()", func() {
        It("gets the apps", func() {
            c := internal.NewCapiClient(func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                Expect(method).To(Equal(http.MethodGet))
                Expect(path).To(And(
                    ContainSubstring("/v3/apps"),
                    ContainSubstring("lemons=limes"),
                    ContainSubstring("mangoes=limes"),
                ))
                return []byte(validAppsResponse), nil
            })

            apps, err := c.Apps(map[string]string{
                "lemons":  "limes",
                "mangoes": "limes",
            })

            Expect(err).ToNot(HaveOccurred())
            Expect(apps).To(ConsistOf(models.App{
                Guid: "app-guid",
            }))
        })

        DescribeTable("errors", func(do func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error)) {
            c := internal.NewCapiClient(do)

            _, err := c.Apps(nil)
            Expect(err).To(HaveOccurred())
        },
            Entry("do returns an error", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return nil, errors.New("expected")
            }),
            Entry("do returns invalid json", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return []byte("{]"), nil
            }),
        )
    })

    Describe("Process()", func() {
        It("gets the process", func() {
            c := internal.NewCapiClient(func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                Expect(method).To(Equal(http.MethodGet))
                Expect(path).To(Equal("/v3/apps/app-guid/processes/process-type"))
                return []byte(validProcessResponse), nil
            })

            process, err := c.Process("app-guid", "process-type")
            Expect(err).ToNot(HaveOccurred())

            Expect(process).To(Equal(models.Process{
                Instances: 2,
            }))
        })

        DescribeTable("errors", func(do func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error)) {
            c := internal.NewCapiClient(do)

            _, err := c.Process("app-guid", "process-type")
            Expect(err).To(HaveOccurred())
        },
            Entry("do returns an error", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return nil, errors.New("expected")
            }),
            Entry("do returns invalid json", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return []byte("{]"), nil
            }),
        )
    })

    Describe("Scale()", func() {
        It("scales the process", func() {
            var called bool
            c := internal.NewCapiClient(func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                called = true
                Expect(method).To(Equal(http.MethodPost))
                Expect(path).To(Equal("/v3/apps/app-guid/processes/process-type/actions/scale"))
                Expect(body).To(MatchJSON(`{ "instances": 5 }`))
                return nil, nil
            })

            err := c.Scale("app-guid", "process-type", 5)
            Expect(err).ToNot(HaveOccurred())
            Expect(called).To(BeTrue())
        })

        DescribeTable("errors", func(do func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error)) {
            c := internal.NewCapiClient(do)
            Expect(c.Scale("app-guid", "process-type", 5)).ToNot(Succeed())
        },
            Entry("do returns an error", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return nil, errors.New("expected")
            }),
        )
    })

    Describe("CreateTask()", func() {
        It("creates a task", func() {
            var called bool
            c := internal.NewCapiClient(func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
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
                return []byte(validTaskResponse), nil
            })

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
            c := internal.NewCapiClient(func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                for _, o := range opts {
                    o(&http.Header{})
                }
                return []byte(validTaskResponse), nil
            })

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

        DescribeTable("errors", func(do func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error)) {
            c := internal.NewCapiClient(do)
            _, err := c.CreateTask("app-guid", "command", models.TaskConfig{})
            Expect(err).To(HaveOccurred())
        },
            Entry("do returns an error", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return nil, errors.New("expected")
            }),
            Entry("do returns invalid json", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return []byte("{]"), nil
            }),
        )
    })

    Describe("Stop()", func() {
        It("stops the process", func() {
            var called bool
            c := internal.NewCapiClient(func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                called = true
                Expect(method).To(Equal(http.MethodPost))
                Expect(path).To(Equal("/v3/apps/app-guid/actions/stop"))
                return nil, nil
            })

            err := c.Stop("app-guid")
            Expect(err).ToNot(HaveOccurred())
            Expect(called).To(BeTrue())
        })

        DescribeTable("errors", func(do func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error)) {
            c := internal.NewCapiClient(do)
            err := c.Stop("app-guid")
            Expect(err).To(HaveOccurred())
        },
            Entry("do returns an error", func(method, path, body string, opts ...internal.HeaderOption) ([]byte, error) {
                return nil, errors.New("expected")
            }),
        )
    })
})

const validAppsResponse = `{"resources": [ { "guid": "app-guid" } ]}`
const validProcessResponse = `{ "instances": 2 }`
const validTaskResponse = `{"guid": "task-guid"}`
