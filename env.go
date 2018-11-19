package client

import (
    "encoding/json"
    "log"
    "time"

    "code.cloudfoundry.org/go-envstruct"
)

type environment struct {
    CloudControllerApi string          `env:"API_URL"`
    HttpTimeout        time.Duration   `env:"HTTP_TIMEOUT"`
    SkipSslValidation  bool            `env:"SKIP_SSL_VALIDATION"`
    VcapApplication    VcapApplication `env:"VCAP_APPLICATION, required"`
    VcapServices       VcapServices    `env:"VCAP_SERVICES, required"`
}

// VcapApplication is information provided by the Cloud Foundry runtime
type VcapApplication struct {
    SpaceID string `json:"space_id"`
    CfApi   string `json:"cf_api"`
}

// UnmarshalEnv decodes a VcapApplication for envstruct.Load
func (v *VcapApplication) UnmarshalEnv(data string) error {
    return json.Unmarshal([]byte(data), v)
}

type userProvided struct {
    InstanceName string            `json:"instance_name"`;
    Credentials  map[string]string `json:"credentials"`
}

// VcapServices is information provided by the Cloud Foundry runtime
type VcapServices struct {
    Credhub []struct {
        Credentials map[string]string `json:"credentials"`
    } `json:"credhub"`
    UserProvided []userProvided `json:"user-provided"`
}

// UnmarshalEnv decodes a VcapServices for envstruct.Load
func (v *VcapServices) UnmarshalEnv(data string) error {
    if data == "" {
        return nil
    }

    return json.Unmarshal([]byte(data), v)
}

// LoadEnv instantiates an Environment via envstruct
func LoadEnv() environment {
    env := environment{
        HttpTimeout: 15 * time.Second,
    }

    err := envstruct.Load(&env)
    if err != nil {
        log.Fatalf("unable to load environment: %s", err)
    }

    if env.CloudControllerApi == "" {
        env.CloudControllerApi = env.VcapApplication.CfApi
    }

    return env
}

func getUserProvidedCredentials(userProvided []userProvided) (map[string]string, bool) {
    for _, service := range userProvided {
        if service.InstanceName == "pvtl_app_automation_credentials" {
            return service.Credentials, true
        }
    }

    return nil, false
}
