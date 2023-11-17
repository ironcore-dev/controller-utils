// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package configutils

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("configutils")
)

// EgressSelectionName is the name of the egress configuration to use.
type EgressSelectionName string

const (
	// EgressSelectionNameControlPlane instructs to use the controlplane egress selection.
	EgressSelectionNameControlPlane EgressSelectionName = "controlplane"
	// EgressSelectionNameEtcd instructs to use the etcd egress selection.
	EgressSelectionNameEtcd EgressSelectionName = "etcd"
	// EgressSelectionNameCluster instructs to use the cluster egress selection.
	EgressSelectionNameCluster EgressSelectionName = "cluster"
)

// NetworkContext returns the corresponding network context of the egress selection.
func (n EgressSelectionName) NetworkContext() (egressselector.NetworkContext, error) {
	switch n {
	case EgressSelectionNameControlPlane:
		return egressselector.ControlPlane.AsNetworkContext(), nil
	case EgressSelectionNameEtcd:
		return egressselector.Etcd.AsNetworkContext(), nil
	case EgressSelectionNameCluster:
		return egressselector.Cluster.AsNetworkContext(), nil
	default:
		return egressselector.NetworkContext{}, fmt.Errorf("unknown egress selection name %q", n)
	}
}

// GetConfigOptions are options to supply for a GetConfig call.
type GetConfigOptions struct {
	// Context is the kubeconfig context to load.
	Context string
	// Kubeconfig is the path to a kubeconfig to load.
	// If unset, the '--kubeconfig' flag is used.
	Kubeconfig *string
	// EgressSelectorConfig is the path to an egress selector config to load.
	EgressSelectorConfig string
	// EgressSelectionName is the name of the egress configuration to use.
	// Defaults to EgressSelectionNameControlPlane.
	EgressSelectionName EgressSelectionName
}

// ApplyToGetConfig implements GetConfigOption.
func (o *GetConfigOptions) ApplyToGetConfig(o2 *GetConfigOptions) {
	if o.Context != "" {
		o2.Context = o.Context
	}
	if o.Kubeconfig != nil {
		o2.Kubeconfig = pointer.String(*o.Kubeconfig)
	}
	if o.EgressSelectorConfig != "" {
		o2.EgressSelectorConfig = o.EgressSelectorConfig
	}
	if o.EgressSelectionName != "" {
		o2.EgressSelectionName = o.EgressSelectionName
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

// EgressSelectorConfig allows specifying the path to an egress selector config to use.
type EgressSelectorConfig string

func (c EgressSelectorConfig) ApplyToGetConfig(o *GetConfigOptions) {
	o.EgressSelectorConfig = string(c)
}

type WithEgressSelectionName EgressSelectionName

// ApplyToGetConfig implements GetConfigOption.
func (w WithEgressSelectionName) ApplyToGetConfig(o *GetConfigOptions) {
	o.EgressSelectionName = EgressSelectionName(w)
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

func setGetConfigOptionsDefaults(o *GetConfigOptions) {
	if o.EgressSelectionName == "" {
		o.EgressSelectionName = EgressSelectionNameControlPlane
	}
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
	setGetConfigOptionsDefaults(o)

	var kubeconfig string
	if o.Kubeconfig != nil {
		kubeconfig = *o.Kubeconfig
	} else {
		kubeconfig = getKubeconfigFlag()
	}

	cfg, err := loadConfig(kubeconfig, o.Context)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	if err := applyEgressSelector(o.EgressSelectorConfig, o.EgressSelectionName, cfg); err != nil {
		return nil, fmt.Errorf("error applying egress selector: %w", err)
	}

	return cfg, nil
}

func loadConfig(kubeconfig, context string) (*rest.Config, error) {
	// If a flag is specified with the config location, use that
	if len(kubeconfig) > 0 {
		return loadConfigWithContext("", &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}, context)
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

	return loadConfigWithContext("", loadingRules, context)
}

func applyEgressSelector(egressSelectorConfig string, egressSelectionName EgressSelectionName, cfg *rest.Config) error {
	if egressSelectorConfig == "" {
		return nil
	}

	networkContext, err := egressSelectionName.NetworkContext()
	if err != nil {
		return fmt.Errorf("error obtaining network context: %w", err)
	}

	egressSelectorCfg, err := egressselector.ReadEgressSelectorConfiguration(egressSelectorConfig)
	if err != nil {
		return fmt.Errorf("error reading egress selector configuration: %w", err)
	}

	egressSelector, err := egressselector.NewEgressSelector(egressSelectorCfg)
	if err != nil {
		return fmt.Errorf("error creating egress selector: %w", err)
	}

	dial, err := egressSelector.Lookup(networkContext)
	if err != nil {
		return fmt.Errorf("error looking up network context %s: %w", networkContext.EgressSelectionName.String(), err)
	}
	if dial == nil {
		return fmt.Errorf("no dialer for network context %s", networkContext.EgressSelectionName.String())
	}

	cfg.Dial = dial
	return nil
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
