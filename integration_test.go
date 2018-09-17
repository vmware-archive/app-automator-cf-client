package client_test

import (
    "crypto/tls"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
    "time"

    "github.com/pivotal-cf/eats-cf-client"
    "github.com/pivotal-cf/eats-cf-client/models"

    "github.com/gorilla/mux"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

const (
    username = "admin"
    password = "supersecret"
    token    = "bearer this-is-my-token"
)

type integrationTestContext struct {
    server            *httptest.Server
    cfg               client.Config
    httpTimeout       time.Duration
    skipSslValidation bool

    oauthCalled    int
    getAppsQuery   url.Values
    getProcessVars map[string]string
    scaleVars      map[string]string
    scaleBody      string

    createTaskVars map[string]string
    createTaskBody string
    requestDelay   time.Duration
}

var _ = Describe("Client Integration", func() {
    Describe("Scale()", func() {
        It("gets app information", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.cfg)

            Expect(c.Scale("lemons", 2)).To(Succeed())

            Expect(tc.getAppsQuery).To(And(
                HaveKeyWithValue("space_guids", []string{"space-guid"}),
            ))
        })

        It("calls the scale process action with the new instance count", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.cfg)
            Expect(c.Scale("lemons", 2)).To(Succeed())

            Expect(tc.scaleVars).To(And(
                HaveKeyWithValue("appGuid", "app-guid"),
                HaveKeyWithValue("processType", "web"),
            ))

            Expect(tc.scaleBody).To(MatchJSON(`{ "instances": 2 }`))
        })
    })

    Describe("Process()", func() {
        It("gets app information", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.cfg)
            _, err := c.Process("lemons", "web")
            Expect(err).ToNot(HaveOccurred())

            Expect(tc.getAppsQuery).To(And(
                HaveKeyWithValue("space_guids", []string{"space-guid"}),
            ))
        })

        It("gets the process", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.cfg)
            _, err := c.Process("lemons", "web")
            Expect(err).ToNot(HaveOccurred())

            Expect(tc.getProcessVars).To(And(
                HaveKeyWithValue("appGuid", "app-guid"),
                HaveKeyWithValue("processType", "web"),
            ))
        })
    })

    Describe("CreateTask()", func() {
        It("gets app information", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.cfg)
            task, err := c.CreateTask("lemons", "command", models.TaskConfig{})
            Expect(err).ToNot(HaveOccurred())

            Expect(tc.getAppsQuery).To(And(
                HaveKeyWithValue("space_guids", []string{"space-guid"}),
            ))
            Expect(task.Guid).To(Equal("task-guid"))
        })

        It("creates the task", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.cfg)
            _, err := c.CreateTask("lemons", "command", models.TaskConfig{
                Name:        "lemons",
                DiskInMB:    7,
                MemoryInMB:  30,
                DropletGUID: "droplet-guid",
            })
            Expect(err).ToNot(HaveOccurred())

            Expect(tc.createTaskVars).To(HaveKeyWithValue("appGuid", "app-guid"))
            Expect(tc.createTaskBody).To(MatchJSON(`{
                "command": "command",
                "name": "lemons",
                "disk_in_mb": 7,
                "memory_in_mb": 30,
                "droplet_guid": "droplet-guid"
            }`))
        })

        It("times out", func() {
            tc, teardown := setupWithTimeout(time.Microsecond)
            defer teardown()

            tc.requestDelay = time.Millisecond
            c := client.New(tc.cfg)
            _, err := c.CreateTask("lemons", "command", models.TaskConfig{})
            Expect(err).To(HaveOccurred())
        })
    })
})

func setup() (*integrationTestContext, func()) {
    return setupWithTimeout(0)
}

func setupWithTimeout(timeout time.Duration) (*integrationTestContext, func()) {
    tc := &integrationTestContext{}

    router := mux.NewRouter()
    tc.server = httptest.NewUnstartedServer(router)
    setupUaa(tc, router)
    setupCc(tc, router)
    tc.server.Start()

    tc.cfg = client.Config{
        CloudControllerUrl: tc.server.URL,
        SpaceGuid:          "space-guid",
        Username:           username,
        Password:           password,
        HttpClient: &http.Client{
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: tc.skipSslValidation,
                },
            },
            Timeout: timeout,
        },
    }

    return tc, func() {
        tc.server.Close()
    }
}

func setupUaa(tc *integrationTestContext, router *mux.Router) {
    router.HandleFunc("/oauth/token", func(w http.ResponseWriter, req *http.Request) {
        tc.oauthCalled++

        w.Header().Set("Content-Type", "application/json")

        tokenPieces := strings.Split(token, " ")
        w.Write([]byte(fmt.Sprintf(`{"access_token": "%s", "token_type": "%s"}`, tokenPieces[1], tokenPieces[0])))
    }).Methods(http.MethodPost)
}

func setupCc(tc *integrationTestContext, router *mux.Router) {
    router.HandleFunc("/v3/apps", handleListApps(tc)).Methods(http.MethodGet)
    router.HandleFunc("/v3/apps/{appGuid}/processes/{processType}", handleGetProcess(tc)).Methods(http.MethodGet)
    router.HandleFunc("/v3/apps/{appGuid}/processes/{processType}/actions/scale", handleScale(tc)).Methods(http.MethodPost)
    router.HandleFunc("/v3/apps/{appGuid}/tasks", handleTask(tc)).Methods(http.MethodPost)
}

func handleListApps(tc *integrationTestContext) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{token}))

        time.Sleep(tc.requestDelay)

        tc.getAppsQuery = req.URL.Query()
        w.Write([]byte(validAppsResponse))
    }
}

func handleGetProcess(tc *integrationTestContext) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{token}))

        time.Sleep(tc.requestDelay)

        tc.getProcessVars = mux.Vars(req)
        w.Write([]byte(validProcessResponse))
    }
}

func handleScale(tc *integrationTestContext) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{token}))

        time.Sleep(tc.requestDelay)

        tc.scaleVars = mux.Vars(req)

        body, err := ioutil.ReadAll(req.Body)
        Expect(err).ToNot(HaveOccurred())
        tc.scaleBody = string(body)

        w.WriteHeader(http.StatusCreated)
    }
}

func handleTask(tc *integrationTestContext) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{token}))

        time.Sleep(tc.requestDelay)

        tc.createTaskVars = mux.Vars(req)

        body, err := ioutil.ReadAll(req.Body)
        Expect(err).ToNot(HaveOccurred())
        tc.createTaskBody = string(body)

        w.WriteHeader(http.StatusCreated)
        w.Write([]byte(validTaskResponse))
    }
}

const validAppsResponse = `{
    "resources": [{
        "name": "lemons",
        "guid": "app-guid"
    }]
}`

const validProcessResponse = `{ "instances": 3 }`

const validTaskResponse = `{"guid": "task-guid"}`
