// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configutils

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ = ginkgo.Describe("Configutils", func() {
	ginkgo.Describe("GetConfig", func() {
		var (
			apiConfig   *clientcmdapi.Config
			config      *rest.Config
			otherConfig *rest.Config
		)
		ginkgo.BeforeEach(func() {
			apiConfig = &clientcmdapi.Config{
				Clusters: map[string]*clientcmdapi.Cluster{
					"default": {
						Server: "http://example.org",
					},
					"other": {
						Server: "http://other.example.org",
					},
				},
				AuthInfos: map[string]*clientcmdapi.AuthInfo{
					"default": {},
					"other":   {},
				},
				Contexts: map[string]*clientcmdapi.Context{
					"default": {
						Cluster:   "default",
						AuthInfo:  "default",
						Namespace: corev1.NamespaceDefault,
					},
					"other": {
						Cluster:   "other",
						AuthInfo:  "other",
						Namespace: corev1.NamespaceDefault,
					},
				},
				CurrentContext: "default",
			}
			config = &rest.Config{
				Host: "http://example.org",
			}
			otherConfig = &rest.Config{
				Host: "http://other.example.org",
			}
		})

		ginkgo.It("should load the config at the kubeconfig path", func() {
			testFile, err := os.CreateTemp("", "kubeconfig")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(testFile.Name())).To(Succeed())
			}()

			Expect(clientcmd.WriteToFile(*apiConfig, testFile.Name())).To(Succeed())
			kubeconfig = testFile.Name()

			loaded, err := GetConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded).To(Equal(config))
		})

		ginkgo.It("should load the config at the specified kubeconfig option", func() {
			testFile, err := os.CreateTemp("", "kubeconfig")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(testFile.Name())).To(Succeed())
			}()

			Expect(clientcmd.WriteToFile(*apiConfig, testFile.Name())).To(Succeed())

			loaded, err := GetConfig(Kubeconfig(testFile.Name()))
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded).To(Equal(config))
		})

		ginkgo.It("should load the config with the given context", func() {
			testFile, err := os.CreateTemp("", "kubeconfig")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(testFile.Name())).To(Succeed())
			}()

			Expect(clientcmd.WriteToFile(*apiConfig, testFile.Name())).To(Succeed())
			kubeconfig = testFile.Name()

			loaded, err := GetConfig(Context("other"))
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded).To(Equal(otherConfig))
		})

		ginkgo.It("should error if the kubeconfig does not exist", func() {
			kubeconfig = "should definitely not exist - ever"
			_, err := GetConfig()
			Expect(err).To(HaveOccurred())
		})
	})
})
