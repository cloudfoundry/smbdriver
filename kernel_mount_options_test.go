package smbdriver_test

import (
	"code.cloudfoundry.org/smbdriver"
	"fmt"
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KernelMountOptions", func() {
	Describe("#ToKernelMountOptionString", func() {
		var (
			mountOpts          map[string]interface{}
			kernelMountOptions string
		)

		BeforeEach(func() {
			mountOpts = make(map[string]interface{})
		})

		JustBeforeEach(func() {
			kernelMountOptions = smbdriver.ToKernelMountOptionString(mountOpts)
		})

		Context("given an empty mount opts", func() {
			It("should return an empty mount opts string", func() {
				Expect(kernelMountOptions).To(BeEmpty())
			})
		})

		Context("given a mount opts", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"opt1": "val1",
					"opt2": "val2",
				}
			})

			It("should return a valid mount opts string", func() {
				Expect(kernelMountOptions).To(Equal("opt1=val1,opt2=val2"))
			})
		})

		Context("given an integer option value with a leading zero", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"opt1": "0123",
				}
			})

			It("strips the leading zero from the mount option string", func() {
				Expect(kernelMountOptions).To(Equal("opt1=123"))
			})
		})

		Context("given an integer option value", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"opt1": math.MaxInt64,
				}
			})

			It("strips the leading zero from the mount option string", func() {
				Expect(kernelMountOptions).To(Equal(fmt.Sprintf("opt1=%d", math.MaxInt64)))
			})
		})

		Context("given a mount option with no value", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"does-not-matter": "",
				}
			})

			It("contains the mount option", func() {
				Expect(kernelMountOptions).To(ContainSubstring("does-not-matter"))
			})
		})

		Context("given a mount option with nil value", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"does-not-matter": nil,
				}
			})

			It("omits the mount option", func() {
				Expect(kernelMountOptions).NotTo(ContainSubstring("does-not-matter"))
			})
		})

		Context("given a 'Domain' mount option with no value", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"domain": "",
					"Domain": "",
				}
			})

			It("omits the mount option", func() {
				Expect(kernelMountOptions).NotTo(ContainSubstring("domain"))
				Expect(kernelMountOptions).NotTo(ContainSubstring("Domain"))
			})
		})

		Context("given a 'Domain' mount option with nil value", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"domain": nil,
				}
			})

			It("omits the mount option", func() {
				Expect(kernelMountOptions).NotTo(ContainSubstring("domain"))
			})
		})

		Context("given a readonly mount option with a string boolean value", func() {
			BeforeEach(func() {
				mountOpts = map[string]interface{}{
					"ro": "true",
				}
			})

			It("includes the mount option", func() {
				Expect(kernelMountOptions).To(ContainSubstring("ro"))
			})
		})
	})

})
