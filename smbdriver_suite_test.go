package smbdriver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

	"testing"
)

func TestSMBDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SMBDriver Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Minute)
})