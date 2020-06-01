package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTerratag(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terratag Suite")
}
