package client

import (
    "encoding/json"
    "log"
    "time"

    "code.cloudfoundry.org/go-envstruct"
)

// Environment holds configuration
type Environment struct {
    CloudControllerApi string          `env:"API_URL"`
    HttpTimeout        time.Duration   `env:"HTTP_TIMEOUT"`
    SkipSslValidation  bool            `env:"SKIP_SSL_VALIDATION"`
    VcapApplication    VcapApplication `env:"VCAP_APPLICATION, required"`
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

// VcapServices is information provided by the Cloud Foundry runtime
type VcapServices struct {
    Credhub []struct {
        Credentials map[string]string `json:"credentials"`
    } `json:"credhub"`
}

// UnmarshalEnv decodes a VcapServices for envstruct.Load
func (v *VcapServices) UnmarshalEnv(data string) error {
    if data == "" {
        return nil
    }

    return json.Unmarshal([]byte(data), v)
}

// LoadEnv instantiates an Environment via envstruct
func LoadEnv() Environment {
    env := Environment{
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
