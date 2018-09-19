package main_test

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	var (
		session *gexec.Session
		command *exec.Cmd
		dir     string
	)

	BeforeEach(func() {
		command = exec.Command(driverPath)
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		session.Kill().Wait()
	})

	Context("with a driver path", func() {
		BeforeEach(func() {
			var err error
			dir, err = ioutil.TempDir("", "driversPath")
			Expect(err).ToNot(HaveOccurred())

			command.Args = append(command.Args, "-driversPath="+dir)
			command.Args = append(command.Args, "-transport=tcp-json")
		})

		It("listens on tcp/8589 by default", func() {
			EventuallyWithOffset(1, func() error {
				_, err := net.Dial("tcp", "127.0.0.1:8589")
				return err
			}, 5).ShouldNot(HaveOccurred())

			specFile := filepath.Join(dir, "smbdriver.json")
			specFileContents, err := ioutil.ReadFile(specFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(specFileContents)).To(MatchJSON(`{
				"Name": "smbdriver",
				"Addr": "http://127.0.0.1:8589",
				"TLSConfig": null,
				"UniqueVolumeIds": false
			}`))
		})

		Context("with unique volume IDs enabled", func() {
			BeforeEach(func() {
				command.Args = append(command.Args, "-uniqueVolumeIds")
			})

			It("sets the uniqueVolumeIds flag in the spec file", func() {
				specFile := filepath.Join(dir, "smbdriver.json")
				Eventually(func() error {
					_, err := os.Stat(specFile)
					return err
				}, 5).ShouldNot(HaveOccurred())

				specFileContents, err := ioutil.ReadFile(specFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(specFileContents)).To(MatchJSON(`{
					"Name": "smbdriver",
					"Addr": "http://127.0.0.1:8589",
					"TLSConfig": null,
					"UniqueVolumeIds": true
				}`))
			})
		})
	})
})
