package client_test

import (
    "testing"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

func TestEatsCfClient(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "EatsCfClient Suite")
}
