// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/utils/containerd"
)

type mutator struct {
	client client.Client
	config containerd.Configuration
}

func (m *mutator) Mutate(ctx context.Context, new, _ client.Object) error {
	log := logf.FromContext(ctx)

	cluster, err := extensionscontroller.GetCluster(ctx, m.client, new.GetNamespace())
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	osc, ok := new.(*extensionsv1alpha1.OperatingSystemConfig)
	if !ok {
		return fmt.Errorf("expected new object to be of type *extensionsv1alpha1.OperatingSystemConfig, got %T", new)
	}

	if osc.Spec.CRIConfig == nil || osc.Spec.CRIConfig.Name != extensionsv1alpha1.CRINameContainerD {
		return nil
	}

	var (
		shootProvider = cluster.Shoot.Spec.Provider.Type
		shootRegion   = cluster.Shoot.Spec.Region
	)

	switch osc.Spec.Purpose {
	case extensionsv1alpha1.OperatingSystemConfigPurposeReconcile:
		for _, upstreamConfig := range m.config.GetUpstreamConfig(shootProvider, shootRegion) {
			if osc.Spec.CRIConfig.Containerd == nil {
				osc.Spec.CRIConfig.Containerd = &extensionsv1alpha1.ContainerdConfig{}
			}

			// Don't overwrite existing upstream configuration to not collide with other extensions (e.g. registry-cache)
			if hasUpstreamConfiguration(osc.Spec.CRIConfig.Containerd, upstreamConfig.Upstream) {
				continue
			}

			log.V(2).Info("Adding registry mirror configuration for node reconciliation", "upstream", upstreamConfig.Upstream)

			osc.Spec.CRIConfig.Containerd.Registries = append(osc.Spec.CRIConfig.Containerd.Registries, extensionsv1alpha1.RegistryConfig{
				Upstream: upstreamConfig.Upstream,
				Server:   ptr.To(upstreamConfig.Server),
				Hosts:    []extensionsv1alpha1.RegistryHost{{URL: upstreamConfig.HostURL, Capabilities: []extensionsv1alpha1.RegistryCapability{extensionsv1alpha1.PullCapability, extensionsv1alpha1.ResolveCapability}}},
			})
		}

	case extensionsv1alpha1.OperatingSystemConfigPurposeProvision:
		for _, upstreamConfig := range m.config.GetUpstreamConfig(shootProvider, shootRegion) {
			mirror := containerd.RegistryMirror{
				UpstreamServer: upstreamConfig.Server,
				MirrorHost:     upstreamConfig.HostURL,
			}

			log.V(2).Info("Adding registry mirror configuration for node provisioning", "upstream", upstreamConfig.Upstream)

			data, err := mirror.HostsTOML()
			if err != nil {
				return fmt.Errorf("failed to create hosts.toml file for upstream %q: %w", upstreamConfig.Upstream, err)
			}

			osc.Spec.Files = extensionswebhook.EnsureFileWithPath(osc.Spec.Files, extensionsv1alpha1.File{
				Path:        filepath.Join("/etc/containerd/certs.d", upstreamConfig.Upstream, "hosts.toml"),
				Permissions: ptr.To[uint32](0644),
				Content: extensionsv1alpha1.FileContent{
					Inline: &extensionsv1alpha1.FileContentInline{
						Data: data,
					},
				},
			})
		}
	}

	return nil
}

func hasUpstreamConfiguration(containerdConfig *extensionsv1alpha1.ContainerdConfig, upstream string) bool {
	if containerdConfig == nil || containerdConfig.Registries == nil {
		return false
	}

	return slices.ContainsFunc(containerdConfig.Registries, func(registry extensionsv1alpha1.RegistryConfig) bool {
		return registry.Upstream == upstream
	})
}

// NewMutator creates a new Mutator instance.
func NewMutator(client client.Client, config *v1alpha1.Configuration) extensionswebhook.Mutator {
	return &mutator{
		client: client,
		config: containerd.NewConfiguration(config),
	}
}
