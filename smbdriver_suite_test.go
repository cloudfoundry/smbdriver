package smbdriver_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestSMBDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SMBDriver Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Minute)
})
