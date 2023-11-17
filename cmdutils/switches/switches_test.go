// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package switches

import (
	"flag"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ = Describe("CMD Switches", func() {
	Context("Testing Switches interface", func() {
		Describe("Enabled", func() {
			It("should return whether an item is enabled or not", func() {
				s := New("runner-a", "runner-b")
				Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())

				Expect(s.Enabled("runner-a")).To(BeTrue())
				Expect(s.Enabled("runner-b")).To(BeFalse())
			})
		})

		Describe("AllEnabled", func() {
			It("should return true when all given switches are enabled", func() {
				s := New("runner-a", Disable("runner-b"), "runner-c")
				Expect(s.Set("*")).NotTo(HaveOccurred())

				Expect(s.AllEnabled("runner-a", "runner-c")).To(BeTrue())
			})

			It("should return false if any of the given switches is not enabled", func() {
				s := New("runner-a", Disable("runner-b"), "runner-c")
				Expect(s.Set("*")).NotTo(HaveOccurred())

				Expect(s.AllEnabled("runner-a", "runner-b")).To(BeFalse())
			})

			It("should return true on empty arguments", func() {
				s := New("runner-a", Disable("runner-b"), "runner-c")
				Expect(s.Set("*")).NotTo(HaveOccurred())

				Expect(s.AllEnabled()).To(BeTrue())
			})
		})

		Describe("AnyEnabled", func() {
			It("should return true when any of the given switches is enabled", func() {
				s := New("runner-a", Disable("runner-b"), "runner-c")
				Expect(s.Set("*")).NotTo(HaveOccurred())

				Expect(s.AnyEnabled("runner-a", "runner-b")).To(BeTrue())
			})

			It("should return false if all of the given switches are not enabled", func() {
				s := New("runner-a", Disable("runner-b"), "runner-c")
				Expect(s.Set("*")).NotTo(HaveOccurred())

				Expect(s.AnyEnabled("runner-b")).To(BeFalse())
			})

			It("should return false on empty arguments", func() {
				s := New("runner-a", Disable("runner-b"), "runner-c")
				Expect(s.Set("*")).NotTo(HaveOccurred())

				Expect(s.AnyEnabled()).To(BeFalse())
			})
		})

		Describe("Values", func() {
			It("should return the switches with their values", func() {
				s := New("runner-a", "runner-b")
				Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())

				Expect(s.Values()).To(Equal(map[string]bool{
					"runner-a": true,
					"runner-b": false,
				}))
			})
		})

		Describe("All", func() {
			It("should return all items", func() {
				s := New("runner-a", "runner-b", Disable("runner-c"))
				Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())

				Expect(s.All()).To(Equal(sets.New("runner-a", "runner-b", "runner-c")))
			})
		})

		Describe("Active", func() {
			It("should return all enabled items", func() {
				s := New("runner-a", "runner-b", Disable("runner-c"))
				Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())

				Expect(s.Active()).To(Equal(sets.New("runner-a")))
			})
		})

		Describe("EnabledByDefault", func() {
			It("should return all items that are enabled by default", func() {
				s := New("runner-a", Disable("runner-b"))

				Expect(s.EnabledByDefault()).To(Equal(sets.New("runner-a")))
			})
		})

		Describe("DisabledByDefault", func() {
			It("should return all items disabled by default", func() {
				s := New("runner-a", Disable("runner-b"))

				Expect(s.DisabledByDefault()).To(Equal(sets.New("runner-b")))
			})
		})

		Describe("String", func() {
			It("should return a string repressentation of the switches", func() {
				s := New("runner-a", "runner-b")
				Expect(s.Set("*,-runner-b")).ToNot(HaveOccurred())

				Expect(s.String()).To(Equal("runner-a,-runner-b"))
			})
		})
	})

	Describe("goflag.Parse", func() {
		It("should disable all controllers when no flag is passed", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": false,
				"runner-b": false,
				"runner-c": false,
			}))
		})

		It("should keep default settings when * is passed", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*"})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": true,
				"runner-b": false,
				"runner-c": true,
			}))
		})

		It("should override default settings", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=runner-a,-runner-c"})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": true,
				"runner-b": false,
				"runner-c": false,
			}))
		})

		It("should override some of default settings", func() {
			fs := flag.NewFlagSet("", flag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*,-runner-a"})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": false,
				"runner-b": false,
				"runner-c": true,
			}))
		})
	})

	Describe("pflag.parse", func() {
		It("should disable all controllers when no flag is passed", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": false,
				"runner-b": false,
				"runner-c": false,
			}))
		})

		It("should keep default settings when * is passed", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*"})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": true,
				"runner-b": false,
				"runner-c": true,
			}))
		})

		It("should override default settings", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=runner-a,-runner-c"})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": true,
				"runner-b": false,
				"runner-c": false,
			}))
		})

		It("should override some of default settings", func() {
			fs := pflag.NewFlagSet("", pflag.ExitOnError)
			controllers := New("runner-a", Disable("runner-b"), "runner-c")
			fs.Var(controllers, "controllers", "")

			Expect(fs.Parse([]string{"--controllers=*,-runner-a"})).NotTo(HaveOccurred())

			Expect(controllers.Values()).To(Equal(map[string]bool{
				"runner-a": false,
				"runner-b": false,
				"runner-c": true,
			}))
		})
	})
})
