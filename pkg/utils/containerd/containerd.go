// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd

import (
	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
)

// UpStreamConfiguration contains the upstream configuration for containerd.
type UpStreamConfiguration struct {
	Upstream string
	Server   string
	HostURL  string
}

// Configuration defines the interface for operating on image configurations.
type Configuration interface {
	// GetProvisionUpstreamConfig returns the containerd upstream configuration for node provision based on provider and region.
	GetProvisionUpstreamConfig(provider string, region string) []UpStreamConfiguration
	// GetReconcileUpstreamConfig returns the containerd upstream configuration for node reconciliation based on provider and region.
	GetReconcileUpstreamConfig(provider string, region string) []UpStreamConfiguration
}

type configuration struct {
	provision []upstreamConfig
	reconcile []upstreamConfig
}

type upstreamConfig struct {
	upstream        string
	server          string
	providerToHosts map[string]hostConfig
}

type hostConfig struct {
	regionToURL map[string]string
}

// GetProvisionUpstreamConfig returns the containerd upstream configuration for node provision based on provider and region.
func (c *configuration) GetProvisionUpstreamConfig(provider string, region string) []UpStreamConfiguration {
	result := make([]UpStreamConfiguration, 0, len(c.provision))

	for _, provision := range c.provision {
		if hosts, providerExists := provision.providerToHosts[provider]; providerExists {
			if hostURL, regionExists := hosts.regionToURL[region]; regionExists {
				result = append(result, UpStreamConfiguration{
					Upstream: provision.upstream,
					Server:   provision.server,
					HostURL:  hostURL,
				})
			}
		}
	}

	return result
}

// GetReconcileUpstreamConfig returns the containerd upstream configuration for node reconciliation based on provider and region.
func (c *configuration) GetReconcileUpstreamConfig(provider string, region string) []UpStreamConfiguration {
	result := make([]UpStreamConfiguration, 0, len(c.reconcile))

	for _, provision := range c.reconcile {
		if hosts, providerExists := provision.providerToHosts[provider]; providerExists {
			if hostURL, regionExists := hosts.regionToURL[region]; regionExists {
				result = append(result, UpStreamConfiguration{
					Upstream: provision.upstream,
					Server:   provision.server,
					HostURL:  hostURL,
				})
			}
		}
	}

	return result
}

// NewConfiguration creates a new containerd configuration from the given configuration.
func NewConfiguration(config *v1alpha1.Configuration) Configuration {
	if config.Containerd == nil {
		return &configuration{}
	}

	conf := &configuration{
		provision: make([]upstreamConfig, 0, len(config.Containerd.Provision)),
		reconcile: make([]upstreamConfig, 0, len(config.Containerd.Provision)),
	}

	for _, provision := range config.Containerd.Provision {
		conf.provision = append(conf.provision, createUpstreamConfig(provision))
	}

	for _, reconcile := range config.Containerd.Reconcile {
		conf.reconcile = append(conf.reconcile, createUpstreamConfig(reconcile))
	}

	return conf
}

func createUpstreamConfig(containerdUpstreamConfig v1alpha1.ContainerdUpstreamConfig) upstreamConfig {
	upstream := upstreamConfig{
		upstream: containerdUpstreamConfig.Upstream,
		server:   containerdUpstreamConfig.Server,
	}

	for _, host := range containerdUpstreamConfig.Hosts {
		if upstream.providerToHosts == nil {
			upstream.providerToHosts = make(map[string]hostConfig)
		}

		for _, region := range host.Regions {
			if upstream.providerToHosts[host.Provider].regionToURL == nil {
				upstream.providerToHosts[host.Provider] = hostConfig{
					regionToURL: make(map[string]string),
				}
			}

			upstream.providerToHosts[host.Provider].regionToURL[region] = host.URL
		}
	}

	return upstream
}
