// +build windows

package smbdriver_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"code.cloudfoundry.org/goshims/ioutilshim/ioutil_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/smbdriver"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
	"code.cloudfoundry.org/volumedriver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SmbMounter", func() {
	var (
		logger      lager.Logger
		testContext context.Context
		env         voldriver.Env
		err         error

		fakeInvoker *voldriverfakes.FakeInvoker
		fakeIoutil  *ioutil_fake.FakeIoutil
		fakeOs      *os_fake.FakeOs

		subject volumedriver.Mounter

		opts map[string]interface{}
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("smb-mounter")
		testContext = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, testContext)
		opts = map[string]interface{}{}

		fakeInvoker = &voldriverfakes.FakeInvoker{}
		fakeIoutil = &ioutil_fake.FakeIoutil{}
		fakeOs = &os_fake.FakeOs{}

		config := smbdriver.NewSmbConfig()
		_ = config.ReadConf("username,password", "", []string{})

		subject = smbdriver.NewSmbMounter(fakeInvoker, fakeOs, fakeIoutil, config)
	})

	Context("#Mount", func() {
		BeforeEach(func() {
			opts["username"] = "fakeusername"
			opts["password"] = "fakepassword"
		})
		JustBeforeEach(func() {
			err = subject.Mount(env, "source", "target", opts)
		})
		Context("when mount succeeds", func() {

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should call the powershell mounter script with the correct arguments", func() {
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("powershell.exe"))
				Expect(args[0]).To(Equal("-file"))
				Expect(args[1]).To(Equal("C:/var/vcap/jobs/smbdriver-windows/scripts/mounter.ps1"))
				Expect(args[2]).To(Equal("-username"))
				Expect(args[3]).To(Equal("fakeusername"))
				Expect(args[4]).To(Equal("-password"))
				Expect(args[5]).To(Equal("fakepassword"))
				Expect(args[6]).To(Equal("-remotePath"))
				Expect(args[7]).To(Equal("source"))
			})

			It("should ensure the target does not exist", func() {
				Expect(fakeOs.RemoveCallCount()).To(Equal(1))
			})

			It("should make a symbolic link", func() {
				Expect(err).NotTo(HaveOccurred())
				_, cmd, args := fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("cmd"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("/c"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("mklink"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("/d"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("target"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("source"))
			})
		})

		Context("when mount errors", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
			})
			It("should return the error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when Remove fails", func() {
			BeforeEach(func() {
				fakeOs.RemoveReturns(errors.New("remove-failed"))
			})
			It("should return the remove error", func() {
				Expect(err).To(MatchError("remove-failed"))
			})
		})

		Context("when error occurs", func() {
			BeforeEach(func() {
				opts = map[string]interface{}{}

				config := smbdriver.NewSmbConfig()
				_ = config.ReadConf("password", "", []string{"username"})

				subject = smbdriver.NewSmbMounter(fakeInvoker, fakeOs, fakeIoutil, config)

				fakeInvoker.InvokeReturns(nil, nil)
			})

			Context("when a required option is missing", func() {
				It("should error", func() {
					Expect(err).To(MatchError("Missing mandatory options : username"))
				})
			})

			Context("when a disallowed option is passed", func() {
				BeforeEach(func() {
					opts["uid"] = "uid"
				})

				It("should error", func() {
					Expect(err).To(MatchError("Not allowed options : uid"))
				})
			})
		})
	})

	Context("#Unmount", func() {
		Context("when mount succeeds", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns(nil, nil)
				fakeOs.ReadlinkReturns("source", nil)
				err = subject.Unmount(env, "target")
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should call the powershell unmounter script with the right parameters", func() {
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("powershell.exe"))
				Expect(args[0]).To(Equal("-file"))
				Expect(args[1]).To(Equal("C:/var/vcap/jobs/smbdriver-windows/scripts/unmounter.ps1"))
				Expect(args[2]).To(Equal("-remotePath"))
				Expect(args[3]).To(Equal("source"))
			})
		})

		Context("when unmount fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
				err = subject.Unmount(env, "target")
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("#Check", func() {
		var (
			success bool
		)

		Context("when check succeeds", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns(nil, nil)
				success = subject.Check(env, "target", "source")
			})

			It("should use the passed in variables", func() {
				_, cmd, args := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("powershell.exe"))
				Expect(args[0]).To(Equal("-file"))
				Expect(args[1]).To(Equal("C:/var/vcap/jobs/smbdriver-windows/scripts/check_mount.ps1"))
				Expect(args[2]).To(Equal("-remotePath"))
				Expect(args[3]).To(Equal("source"))
			})

			It("uses correct context", func() {
				env, _, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(fmt.Sprintf("%#v", env.Context())).To(ContainSubstring("timerCtx"))
			})

			It("reports valid mountpoint", func() {
				Expect(success).To(BeTrue())
			})
		})

		Context("when check fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("error"))
				success = subject.Check(env, "target", "source")
			})
			It("reports invalid mountpoint", func() {
				Expect(success).To(BeFalse())
			})
		})
	})

	Context("#Purge", func() {
		var (
			rootPath string
		)

		BeforeEach(func() {
			rootPath = filepath.Join("var", "vcap", "data", "some", "path")
		})

		JustBeforeEach(func() {
			subject.Purge(env, rootPath)
		})

		Context("when stuff is in the directory", func() {
			var fakeStuff *ioutil_fake.FakeFileInfo
			BeforeEach(func() {
				fakeStuff = &ioutil_fake.FakeFileInfo{}
				fakeStuff.NameReturns("guidy-guid-guid")
				fakeStuff.IsDirReturns(true)
				fakeIoutil.ReadDirReturns([]os.FileInfo{fakeStuff}, nil)
			})

			It("should remove stuff", func() {
				Expect(fakeOs.RemoveCallCount()).NotTo(BeZero())
				path := fakeOs.RemoveArgsForCall(0)
				Expect(path).To(Equal(filepath.Join(rootPath, "guidy-guid-guid")))
			})

			Context("when the stuff is not a directory", func() {
				BeforeEach(func() {
					fakeStuff.IsDirReturns(false)
				})
				It("should not remove the stuff", func() {
					Expect(fakeOs.RemoveCallCount()).To(BeZero())
				})
			})
		})
	})
})
