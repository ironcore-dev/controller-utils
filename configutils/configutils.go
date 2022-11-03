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
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("configutils")
)

// GetConfigOptions are options to supply for a GetConfig call.
type GetConfigOptions struct {
	// Context is the kubeconfig context to load.
	Context string
	// Kubeconfig is the path to a kubeconfig to load.
	// If unset, the '--kubeconfig' flag is used.
	Kubeconfig *string
}

// ApplyToGetConfig implements GetConfigOption.
func (o *GetConfigOptions) ApplyToGetConfig(o2 *GetConfigOptions) {
	if o.Context != "" {
		o2.Context = o.Context
	}
	if o.Kubeconfig != nil {
		o2.Kubeconfig = pointer.String(*o.Kubeconfig)
	}
}

// ApplyOptions applies all GetConfigOption tro this GetConfigOptions.
func (o *GetConfigOptions) ApplyOptions(opts []GetConfigOption) {
	for _, opt := range opts {
		opt.ApplyToGetConfig(o)
	}
}

// Kubeconfig allows specifying the path to a kubeconfig file to use.
type Kubeconfig string

// ApplyToGetConfig implements GetConfigOption.
func (k Kubeconfig) ApplyToGetConfig(o *GetConfigOptions) {
	o.Kubeconfig = (*string)(&k)
}

// Context allows specifying the context to load.
type Context string

// ApplyToGetConfig implements GetConfigOption.
func (c Context) ApplyToGetConfig(o *GetConfigOptions) {
	o.Context = string(c)
}

// GetConfigOption are options to a GetConfig call.
type GetConfigOption interface {
	// ApplyToGetConfig modifies the underlying GetConfigOptions.
	ApplyToGetConfig(o *GetConfigOptions)
}

func loadConfigWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: context,
		}).ClientConfig()
}

// loadInClusterConfig is a function used to load the in-cluster
// Kubernetes client config. This variable makes is possible to
// test the precedence of loading the config.
var loadInClusterConfig = rest.InClusterConfig

func getKubeconfigFlag() string {
	f := flag.CommandLine.Lookup("kubeconfig")
	if f == nil {
		panic("--kubeconfig flag is not defined")
	}

	return f.Value.String()
}

// GetConfig creates a *rest.Config for talking to a Kubernetes API server.
// Kubeconfig / the '--kubeconfig' flag instruct to use the kubeconfig file at that location.
// Otherwise, will assume running in cluster and use the cluster provided kubeconfig.
//
// It also applies saner defaults for QPS and burst based on the Kubernetes
// controller manager defaults (20 QPS, 30 burst)
//
// # Config precedence
//
// * Kubeconfig / --kubeconfig value / flag pointing at a file
//
// * KUBECONFIG environment variable pointing at a file
//
// * In-cluster config if running in cluster
//
// * $HOME/.kube/config if exists.
func GetConfig(opts ...GetConfigOption) (*rest.Config, error) {
	o := &GetConfigOptions{}
	o.ApplyOptions(opts)

	var kubeconfig string
	if o.Kubeconfig != nil {
		kubeconfig = *o.Kubeconfig
	} else {
		kubeconfig = getKubeconfigFlag()
	}

	// If a flag is specified with the config location, use that
	if len(kubeconfig) > 0 {
		return loadConfigWithContext("", &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}, o.Context)
	}

	// If the recommended kubeconfig env variable is not specified,
	// try the in-cluster config.
	kubeconfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(kubeconfigPath) == 0 {
		if c, err := loadInClusterConfig(); err == nil {
			return c, nil
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	return loadConfigWithContext("", loadingRules, o.Context)
}

// GetConfigOrDie creates a *rest.Config for talking to a Kubernetes apiserver.
// If Kubeconfig / --kubeconfig is set, will use the kubeconfig file at that location. Otherwise, will assume running
// in cluster and use the cluster provided kubeconfig.
//
// Will log an error and exit if there is an error creating the rest.Config.
func GetConfigOrDie(opts ...GetConfigOption) *rest.Config {
	config, err := GetConfig(opts...)
	if err != nil {
		log.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}
	return config
}
