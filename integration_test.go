package client_test

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "strings"

    "github.com/pivotal-cf/eats-cf-client"

    "github.com/gorilla/mux"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "net/url"
)

const (
    username = "admin"
    password = "supersecret"
    token    = "bearer this-is-my-token"
)

type integrationTestContext struct {
    server *httptest.Server
    env       client.Environment

    infoCalled     int
    oauthCalled    int
    getAppsQuery   url.Values
    getProcessVars map[string]string
    scaleVars      map[string]string
    scaleBody string
}

var _ = Describe("Client Integration", func() {
    Describe("Scale()", func() {
        It("gets app information", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.env, username, password)
            Expect(c.Scale("lemons", 2)).To(Succeed())

            Expect(tc.getAppsQuery).To(And(
                HaveKeyWithValue("names", []string{"lemons"}),
                HaveKeyWithValue("space_guids", []string{"space-guid"}),
            ))
        })

        It("gets process information", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.env, username, password)
            Expect(c.Scale("lemons", 2)).To(Succeed())

            Expect(tc.getProcessVars).To(And(
                HaveKeyWithValue("appGuid", "app-guid"),
                HaveKeyWithValue("processType", "web"),
            ))
        })

        It("calls the scale process action with the new instance count", func() {
            tc, teardown := setup()
            defer teardown()

            c := client.New(tc.env, username, password)
            Expect(c.Scale("lemons", 2)).To(Succeed())

            Expect(tc.scaleVars).To(And(
                HaveKeyWithValue("appGuid", "app-guid"),
                HaveKeyWithValue("processType", "web"),
            ))

            Expect(tc.scaleBody).To(MatchJSON(`{ "instances": 5 }`))
        })
    })
})

func setup() (*integrationTestContext, func()) {
    tc := &integrationTestContext{}

    router := mux.NewRouter()
    tc.server = httptest.NewUnstartedServer(router)
    setupUaa(tc, router)
    setupCc(tc, router)
    tc.server.Start()

    tc.env = client.Environment{
        CloudControllerApi: tc.server.URL,
        VcapApplication: client.VcapApplication{
            SpaceID: "space-guid",
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
}

func handleListApps(tc *integrationTestContext) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer this-is-my-token"}))

        tc.getAppsQuery = req.URL.Query()
        w.Write([]byte(validAppsResponse))
    }
}

func handleGetProcess(tc *integrationTestContext) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer this-is-my-token"}))

        tc.getProcessVars = mux.Vars(req)
        w.Write([]byte(validProcessResponse))
    }
}

func handleScale(tc *integrationTestContext) http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{token}))

        tc.scaleVars = mux.Vars(req)

        body, err := ioutil.ReadAll(req.Body)
        Expect(err).ToNot(HaveOccurred())
        tc.scaleBody = string(body)

        w.WriteHeader(http.StatusCreated)
    }
}

const validAppsResponse = `{
    "resources": [{
        "guid": "app-guid"
    }]
}`

const validProcessResponse = `{ "instances": 3 }`
