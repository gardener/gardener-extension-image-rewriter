// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd

import (
	"regexp"

	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
)

// UpStreamConfiguration contains the upstream configuration for containerd.
type UpStreamConfiguration struct {
	Upstream     string
	Server       string
	HostURL      string
	OverridePath *bool
}

// Configuration defines the interface for operating on image configurations.
type Configuration interface {
	// GetUpstreamConfig returns the containerd upstream configuration based on provider and region.
	GetUpstreamConfig(provider string, region string) []UpStreamConfiguration
}

type configuration struct {
	upstreamConfigs []upstreamConfig
}

type upstreamConfig struct {
	upstream        string
	server          string
	providerToHosts map[string]hostConfig
}

type hostConfig struct {
	regionToURL map[string]string
}

var hostWithPathPattern = regexp.MustCompile(`https?://[a-zA-Z0-9\.\-]+(/[^\s]*)+`)

// GetUpstreamConfig returns the containerd upstream configuration based on provider and region.
func (c *configuration) GetUpstreamConfig(provider string, region string) []UpStreamConfiguration {
	result := make([]UpStreamConfiguration, 0, len(c.upstreamConfigs))

	for _, upstreamConf := range c.upstreamConfigs {
		if hosts, providerExists := upstreamConf.providerToHosts[provider]; providerExists {
			if hostURL, regionExists := hosts.regionToURL[region]; regionExists {
				// If the host URL contains a path, override_path needs to be set to true, see https://github.com/containerd/containerd/blob/main/docs/hosts.md#override_path-field.
				var overridePath *bool
				if hostWithPathPattern.MatchString(hostURL) {
					overridePath = ptr.To(true)
				}

				result = append(result, UpStreamConfiguration{
					Upstream:     upstreamConf.upstream,
					Server:       upstreamConf.server,
					HostURL:      hostURL,
					OverridePath: overridePath,
				})
			}
		}
	}

	return result
}

// NewConfiguration creates a new containerd configuration from the given configuration.
func NewConfiguration(config *v1alpha1.Configuration) Configuration {
	conf := &configuration{
		upstreamConfigs: make([]upstreamConfig, 0, len(config.Containerd)),
	}

	for _, containerdConfig := range config.Containerd {
		conf.upstreamConfigs = append(conf.upstreamConfigs, createUpstreamConfig(containerdConfig))
	}

	return conf
}

func createUpstreamConfig(containerdUpstreamConfig v1alpha1.ContainerdConfiguration) upstreamConfig {
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
