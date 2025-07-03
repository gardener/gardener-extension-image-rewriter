// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd_test

import (
	"context"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	. "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/operatingsystemconfig/containerd"
)

var _ = Describe("Mutator", func() {
	var (
		ctx        context.Context
		fakeClient client.Client

		config  *v1alpha1.Configuration
		mutator extensionswebhook.Mutator

		namespace string
		cluster   *extensionsv1alpha1.Cluster
		osc       *extensionsv1alpha1.OperatingSystemConfig
	)

	BeforeEach(func() {
		ctx = context.Background()

		scheme := runtime.NewScheme()
		Expect(extensionsv1alpha1.AddToScheme(scheme)).To(Succeed())
		fakeClient = fakeclient.NewClientBuilder().WithScheme(scheme).Build()

		config = &v1alpha1.Configuration{
			Containerd: &v1alpha1.ContainerdConfiguration{
				Provision: []v1alpha1.ContainerdUpstreamConfig{
					{
						Upstream: "upstream1",
						Server:   "https://server1",
						Hosts: []v1alpha1.ContainerdHostConfig{
							{URL: "https://mirror1-west", Provider: "local", Regions: []string{"west"}},
							{URL: "https://mirror1-central", Provider: "local", Regions: []string{"central", "south", "north"}},
							{URL: "https://mirror1-east", Provider: "local", Regions: []string{"east"}},
						},
					},
					{
						Upstream: "upstream2",
						Server:   "https://server2",
						Hosts: []v1alpha1.ContainerdHostConfig{
							{URL: "https://mirror2-west", Provider: "local2", Regions: []string{"west"}},
							{URL: "https://mirror2-central", Provider: "local2", Regions: []string{"central", "south", "north"}},
							{URL: "https://mirror2-east", Provider: "local2", Regions: []string{"east"}},
						},
					},
				},
				Reconcile: []v1alpha1.ContainerdUpstreamConfig{
					{
						Upstream: "upstream3",
						Server:   "https://server3",
						Hosts: []v1alpha1.ContainerdHostConfig{
							{URL: "https://mirror3-west", Provider: "local2", Regions: []string{"west"}},
							{URL: "https://mirror3-central", Provider: "local2", Regions: []string{"central", "south", "north"}},
							{URL: "https://mirror3-east", Provider: "local2", Regions: []string{"east"}},
						},
					},
					{
						Upstream: "upstream4",
						Server:   "https://server4",
						Hosts: []v1alpha1.ContainerdHostConfig{
							{URL: "https://mirror4-west", Provider: "local", Regions: []string{"west"}},
							{URL: "https://mirror4-central", Provider: "local", Regions: []string{"central", "south", "north"}},
							{URL: "https://mirror4-east", Provider: "local", Regions: []string{"east"}},
						},
					},
				},
			},
		}

		mutator = NewMutator(fakeClient, config)

		namespace = "shoot--test--local"

		cluster = &extensionsv1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
			Spec: extensionsv1alpha1.ClusterSpec{
				Shoot: runtime.RawExtension{
					Raw: []byte(`{
  "apiVersion": "core.gardener.cloud/v1beta1",
  "kind": "Shoot",
  "spec": {
	"provider": {
	  "type": "local"
    },
	"region": "north"
  }
}`),
				},
			},
		}

		Expect(fakeClient.Create(ctx, cluster)).To(Succeed())

		osc = &extensionsv1alpha1.OperatingSystemConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-osc",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
				CRIConfig: &extensionsv1alpha1.CRIConfig{
					Name: extensionsv1alpha1.CRINameContainerD,
				},
			},
		}
	})

	Describe("#Mutate", func() {
		Context("Provision OperatingSystemConfig", func() {
			BeforeEach(func() {
				osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeProvision
			})

			It("should add the containerd configuration as files to the OperatingSystemConfig", func() {
				Expect(mutator.Mutate(ctx, osc, nil)).To(Succeed())

				Expect(osc.Spec.Files).To(ConsistOf(extensionsv1alpha1.File{
					Path:        "/etc/containerd/certs.d/upstream1/hosts.toml",
					Permissions: ptr.To[uint32](0644),
					Content: extensionsv1alpha1.FileContent{
						Inline: &extensionsv1alpha1.FileContentInline{
							Data: `server = "https://server1"

[host."https://mirror1-central"]
  capabilities = ["pull", "resolve"]
`,
						},
					},
				}))
			})

			It("should leave OperatingSystemConfig files unchanged when no configuration matches", func() {
				oscCopy := osc.DeepCopy()
				oscCopy.Namespace = "other-namespace"

				clusterCopy := cluster.DeepCopy()
				clusterCopy.ResourceVersion = ""
				clusterCopy.Name = oscCopy.Namespace
				clusterCopy.Spec.Shoot = runtime.RawExtension{
					Raw: []byte(`{
  "apiVersion": "core.gardener.cloud/v1beta1",
  "kind": "Shoot",
  "spec": {
	"provider": {
	  "type": "local"
    },
	"region": "south-west"
  }
}`),
				}

				Expect(fakeClient.Create(ctx, clusterCopy)).To(Succeed())

				Expect(mutator.Mutate(ctx, oscCopy, nil)).To(Succeed())

				Expect(oscCopy.Spec.Files).To(BeEmpty())
			})
		})

		Context("Reconcile OperatingSystemConfig", func() {
			BeforeEach(func() {
				osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
			})

			It("should add the containerd configuration to the OperatingSystemConfig", func() {
				Expect(mutator.Mutate(ctx, osc, nil)).To(Succeed())

				Expect(osc.Spec.CRIConfig.Containerd.Registries).To(ConsistOf(extensionsv1alpha1.RegistryConfig{
					Upstream: "upstream4",
					Server:   ptr.To("https://server4"),
					Hosts: []extensionsv1alpha1.RegistryHost{
						{URL: "https://mirror4-central", Capabilities: []extensionsv1alpha1.RegistryCapability{extensionsv1alpha1.PullCapability, extensionsv1alpha1.ResolveCapability}},
					},
				}))
			})

			It("should leave already configured upstream unchanged", func() {
				osc.Spec.CRIConfig.Containerd = &extensionsv1alpha1.ContainerdConfig{
					Registries: []extensionsv1alpha1.RegistryConfig{
						{
							Upstream: "upstream4",
							Server:   ptr.To("https://server4"),
							Hosts: []extensionsv1alpha1.RegistryHost{
								{URL: "https://custom-mirror4", Capabilities: []extensionsv1alpha1.RegistryCapability{extensionsv1alpha1.PullCapability, extensionsv1alpha1.ResolveCapability}},
							},
						},
					},
				}

				Expect(mutator.Mutate(ctx, osc, nil)).To(Succeed())

				Expect(osc.Spec.CRIConfig.Containerd.Registries).To(ConsistOf(extensionsv1alpha1.RegistryConfig{
					Upstream: "upstream4",
					Server:   ptr.To("https://server4"),
					Hosts: []extensionsv1alpha1.RegistryHost{
						{URL: "https://custom-mirror4", Capabilities: []extensionsv1alpha1.RegistryCapability{extensionsv1alpha1.PullCapability, extensionsv1alpha1.ResolveCapability}},
					},
				}))
			})

			It("should leave OperatingSystemConfig containerd unchanged when no configuration matches", func() {
				oscCopy := osc.DeepCopy()
				oscCopy.Namespace = "other-namespace"

				clusterCopy := cluster.DeepCopy()
				clusterCopy.ResourceVersion = ""
				clusterCopy.Name = oscCopy.Namespace
				clusterCopy.Spec.Shoot = runtime.RawExtension{
					Raw: []byte(`{
  "apiVersion": "core.gardener.cloud/v1beta1",
  "kind": "Shoot",
  "spec": {
	"provider": {
	  "type": "local"
    },
	"region": "south-west"
  }
}`),
				}

				Expect(fakeClient.Create(ctx, clusterCopy)).To(Succeed())

				Expect(mutator.Mutate(ctx, oscCopy, nil)).To(Succeed())

				Expect(oscCopy.Spec.CRIConfig.Containerd).To(BeNil())
			})
		})
	})
})
