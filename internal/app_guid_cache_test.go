package internal_test

import (
    "errors"

    "github.com/pivotal-cf/eats-cf-client/internal"
    "github.com/pivotal-cf/eats-cf-client/models"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Capi", func() {
    Describe("Get()", func() {
        It("fills cache if not present", func() {
            var appsRefreshed bool
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    Expect(query).To(HaveKeyWithValue("space_guids", "space-guid"))

                    appsRefreshed = true
                    return []models.App{
                        {Name: "limes", Guid: "limes-guid"},
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )

            guid, err := c.Get("lemons")
            Expect(err).ToNot(HaveOccurred())
            Expect(guid).To(Equal("lemons-guid"))
            Expect(appsRefreshed).To(BeTrue())
        })

        It("gets guid from cache if present", func() {
            var appsRefreshed int
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    appsRefreshed++
                    return []models.App{
                        {Name: "limes", Guid: "limes-guid"},
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )

            guid, err := c.Get("lemons")
            Expect(err).ToNot(HaveOccurred())
            Expect(guid).To(Equal("lemons-guid"))

            guid, err = c.Get("lemons")
            Expect(err).ToNot(HaveOccurred())
            Expect(guid).To(Equal("lemons-guid"))

            guid, err = c.Get("limes")
            Expect(err).ToNot(HaveOccurred())
            Expect(guid).To(Equal("limes-guid"))

            Expect(appsRefreshed).To(Equal(1))
        })

        It("handles concurrent reads", func() {
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    return []models.App{
                        {Name: "limes", Guid: "limes-guid"},
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )
            for i := 0; i < 50; i++ {
                go func() { c.Get("lemons") }()
            }
        })

        It("returns an error if the app isn't found", func() {
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    return []models.App{
                        {Name: "limes", Guid: "limes-guid"},
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )

            _, err := c.Get("grapefruit")
            Expect(err).To(HaveOccurred())
        })

        It("returns an error if getting apps fails", func() {
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    return []models.App{
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, errors.New("expected")
                },
                "space-guid",
            )

            _, err := c.Get("lemons")
            Expect(err).To(HaveOccurred())
        })
    })

    Describe("Invalidate()", func() {
        It("clears the cache", func() {
            var appsRefreshed int
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    appsRefreshed++
                    return []models.App{
                        {Name: "limes", Guid: "limes-guid"},
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )

            guid, err := c.Get("lemons")
            Expect(err).ToNot(HaveOccurred())
            Expect(guid).To(Equal("lemons-guid"))

            c.Invalidate()

            guid, err = c.Get("lemons")
            Expect(err).ToNot(HaveOccurred())
            Expect(guid).To(Equal("lemons-guid"))

            Expect(appsRefreshed).To(Equal(2))
        })

        It("handles concurrent reads and invalidations", func() {
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    return []models.App{
                        {Name: "limes", Guid: "limes-guid"},
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )
            for i := 0; i < 50; i++ {
                go func() { c.Get("lemons") }()
            }
            for i := 0; i < 50; i++ {
                go func() { c.Invalidate() }()
            }
        })
    })

    Describe("TryWithRefresh()", func() {
        It("runs the function with the corresponding app guid", func() {
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    return []models.App{
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )

            var called bool
            err := c.TryWithRefresh("lemons", func(appGuid string) error {
                called = true
                Expect(appGuid).To(Equal("lemons-guid"))
                return nil
            })

            Expect(err).ToNot(HaveOccurred())
            Expect(called).To(BeTrue())
        })

        It("retries if the function errors", func() {
            var cacheCallCount int
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    cacheCallCount++
                    if cacheCallCount == 1 {
                        return []models.App{
                            {Name: "lemons", Guid: "wrong-lemons-guid"},
                        }, nil
                    }

                    return []models.App{
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, nil
                },
                "space-guid",
            )

            var appGuids []string
            err := c.TryWithRefresh("lemons", func(appGuid string) error {
                appGuids = append(appGuids, appGuid)
                return errors.New("expected")
            })

            Expect(err).To(HaveOccurred())
            Expect(appGuids).To(ConsistOf("wrong-lemons-guid", "lemons-guid"))
        })

        It("returns an error if app guids can't be fetched", func() {
            c := internal.NewAppGuidCache(
                func(query map[string]string) ([]models.App, error) {
                    return []models.App{
                        {Name: "lemons", Guid: "lemons-guid"},
                    }, errors.New("expected")
                },
                "space-guid",
            )

            err := c.TryWithRefresh("appname", func(appGuid string) error {
                return nil
            })
            Expect(err).To(HaveOccurred())
        })
    })
})
