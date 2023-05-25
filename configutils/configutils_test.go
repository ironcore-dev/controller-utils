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
	"flag"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiserverv1beta1 "k8s.io/apiserver/pkg/apis/apiserver/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

func setKubeconfigFlag(kubeconfig string) {
	ExpectWithOffset(1, flag.CommandLine.Set("kubeconfig", kubeconfig)).To(Succeed())
}

var _ = ginkgo.Describe("Configutils", func() {
	ginkgo.Describe("GetConfig", func() {

		var (
			apiConfig    *clientcmdapi.Config
			egressConfig *apiserverv1beta1.EgressSelectorConfiguration
			config       *rest.Config
			otherConfig  *rest.Config

			configFile       string
			egressConfigFile string
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
			egressConfig = &apiserverv1beta1.EgressSelectorConfiguration{
				TypeMeta: metav1.TypeMeta{
					APIVersion: apiserverv1beta1.SchemeGroupVersion.String(),
					Kind:       "EgressSelectorConfiguration",
				},
				EgressSelections: []apiserverv1beta1.EgressSelection{
					{
						Name: "controlplane",
						Connection: apiserverv1beta1.Connection{
							ProxyProtocol: apiserverv1beta1.ProtocolDirect,
						},
					},
				},
			}

			tempDir := ginkgo.GinkgoT().TempDir()

			configFile = filepath.Join(tempDir, "kubeconfig")
			Expect(clientcmd.WriteToFile(*apiConfig, configFile)).To(Succeed())

			egressConfigData, err := yaml.Marshal(egressConfig)
			Expect(err).NotTo(HaveOccurred())

			egressConfigFile = filepath.Join(tempDir, "egress-config.yaml")
			Expect(os.WriteFile(egressConfigFile, egressConfigData, 0666)).To(Succeed())
		})

		ginkgo.It("should load the config at the kubeconfig path", func() {
			testFile, err := os.CreateTemp("", "kubeconfig")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(testFile.Name())).To(Succeed())
			}()

			Expect(clientcmd.WriteToFile(*apiConfig, testFile.Name())).To(Succeed())
			setKubeconfigFlag(testFile.Name())

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
			setKubeconfigFlag(testFile.Name())

			loaded, err := GetConfig(Context("other"))
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded).To(Equal(otherConfig))
		})

		ginkgo.It("should error if the kubeconfig does not exist", func() {
			setKubeconfigFlag("should definitely not exist - ever")
			_, err := GetConfig()
			Expect(err).To(HaveOccurred())
		})

		ginkgo.It("should load the kubeconfig and apply the egress selector", func() {
			cfg, err := GetConfig(Kubeconfig(configFile), EgressSelectorConfig(egressConfigFile))
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Dial).NotTo(BeNil())
		})

		ginkgo.It("should error if the egress selector config file does not exist", func() {
			_, err := GetConfig(Kubeconfig(configFile), EgressSelectorConfig("should-never-exist"))
			Expect(err).To(HaveOccurred())
		})

		ginkgo.It("should error if the egress context does not exist", func() {
			_, err := GetConfig(Kubeconfig(configFile),
				EgressSelectorConfig(egressConfigFile),
				WithEgressSelectionName(EgressSelectionNameEtcd),
			)
			Expect(err).To(HaveOccurred())
		})
	})
})
