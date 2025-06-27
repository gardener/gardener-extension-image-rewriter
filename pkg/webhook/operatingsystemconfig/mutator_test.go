// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operatingsystemconfig_test

import (
	"context"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	gardenerutils "github.com/gardener/gardener/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	. "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/operatingsystemconfig"
)

var _ = Describe("Mutator", func() {
	var (
		ctx        context.Context
		fakeClient client.Client

		config  *v1alpha1.Configuration
		mutator extensionswebhook.Mutator

		namespace           string
		cluster             *extensionsv1alpha1.Cluster
		nodeAgentInlineData string
		osc                 *extensionsv1alpha1.OperatingSystemConfig
	)

	BeforeEach(func() {
		ctx = context.Background()

		scheme := runtime.NewScheme()
		Expect(extensionsv1alpha1.AddToScheme(scheme)).To(Succeed())
		fakeClient = fakeclient.NewClientBuilder().WithScheme(scheme).Build()

		config = &v1alpha1.Configuration{
			Overwrites: []v1alpha1.ImageOverwrite{
				{
					Source: v1alpha1.Image{Prefix: ptr.To("gardener.cloud/gardener-project")},
					Targets: []v1alpha1.TargetConfiguration{
						{
							Image:    v1alpha1.Image{Prefix: ptr.To("registry.north.local/replicas")},
							Provider: "local",
							Regions:  []string{"north"},
						},
					},
				},
				{
					Source: v1alpha1.Image{Image: ptr.To("sandbox-image:latest")},
					Targets: []v1alpha1.TargetConfiguration{
						{
							Image:    v1alpha1.Image{Image: ptr.To("local-north-sandbox-image:latest")},
							Provider: "local",
							Regions:  []string{"north"},
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

		nodeAgentInlineData = gardenerutils.EncodeBase64([]byte("this is a test for image gardener.cloud/gardener-project/node-agent:latest"))

		osc = &extensionsv1alpha1.OperatingSystemConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-osc",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
				CRIConfig: &extensionsv1alpha1.CRIConfig{
					Name: extensionsv1alpha1.CRINameContainerD,
					Containerd: &extensionsv1alpha1.ContainerdConfig{
						SandboxImage: "sandbox-image:latest",
					},
				},
				Files: []extensionsv1alpha1.File{
					{
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Data:     nodeAgentInlineData,
								Encoding: "b64",
							},
						},
					},
					{
						Content: extensionsv1alpha1.FileContent{
							ImageRef: &extensionsv1alpha1.FileContentImageRef{Image: "gardener.cloud/gardener-project/hyperkube:latest"},
						},
					},
					{
						Content: extensionsv1alpha1.FileContent{
							ImageRef: &extensionsv1alpha1.FileContentImageRef{Image: "gardener.cloud/vali-project/vali:latest"},
						},
					},
				},
			},
		}
	})

	Describe("#Mutate", func() {
		Context("Provisioning OperatingSystemConfig", func() {
			BeforeEach(func() {
				osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeProvision
			})

			It("should mutate all relevant container images", func() {
				nodeAgentInlineDataWithReplacedImage := gardenerutils.EncodeBase64([]byte("this is a test for image registry.north.local/replicas/node-agent:latest"))

				Expect(mutator.Mutate(ctx, osc, nil)).To(Succeed())

				Expect(osc.Spec.Files).To(ConsistOf(
					extensionsv1alpha1.File{
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Data:     nodeAgentInlineDataWithReplacedImage,
								Encoding: "b64",
							},
						},
					},
					extensionsv1alpha1.File{
						Content: extensionsv1alpha1.FileContent{
							ImageRef: &extensionsv1alpha1.FileContentImageRef{Image: "gardener.cloud/gardener-project/hyperkube:latest"},
						},
					},
					extensionsv1alpha1.File{
						Content: extensionsv1alpha1.FileContent{
							ImageRef: &extensionsv1alpha1.FileContentImageRef{Image: "gardener.cloud/vali-project/vali:latest"},
						},
					},
				))

				Expect(osc.Spec.CRIConfig.Containerd.SandboxImage).To(Equal("sandbox-image:latest"))
			})
		})

		Context("Reconciling OperatingSystemConfig", func() {
			BeforeEach(func() {
				osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
			})

			It("should mutate all relevant container images", func() {
				Expect(mutator.Mutate(ctx, osc, nil)).To(Succeed())

				Expect(osc.Spec.Files).To(ConsistOf(
					extensionsv1alpha1.File{
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Data:     nodeAgentInlineData,
								Encoding: "b64",
							},
						},
					},
					extensionsv1alpha1.File{
						Content: extensionsv1alpha1.FileContent{
							ImageRef: &extensionsv1alpha1.FileContentImageRef{Image: "registry.north.local/replicas/hyperkube:latest"},
						},
					},
					extensionsv1alpha1.File{
						Content: extensionsv1alpha1.FileContent{
							ImageRef: &extensionsv1alpha1.FileContentImageRef{Image: "gardener.cloud/vali-project/vali:latest"},
						},
					},
				))

				Expect(osc.Spec.CRIConfig.Containerd.SandboxImage).To(Equal("local-north-sandbox-image:latest"))
			})
		})
	})
})
