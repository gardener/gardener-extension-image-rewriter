// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package pod_test

import (
	"context"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/utils/image"
	. "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/pod"
)

var _ = Describe("Mutator", func() {
	var (
		config      *v1alpha1.Configuration
		imageConfig image.Configuration
		mutator     extensionswebhook.Mutator
	)

	BeforeEach(func() {
		config = &v1alpha1.Configuration{
			Overwrites: []v1alpha1.ImageOverwrite{
				{
					Source: v1alpha1.Image{Image: ptr.To("source-image:latest")},
					Targets: []v1alpha1.TargetConfiguration{
						{
							Image:    v1alpha1.Image{Image: ptr.To("target-image:latest")},
							Provider: "local",
							Regions:  []string{"north"},
						},
					},
				},
				{
					Source: v1alpha1.Image{Image: ptr.To("init-source-image:latest")},
					Targets: []v1alpha1.TargetConfiguration{
						{
							Image:    v1alpha1.Image{Image: ptr.To("init-target-image:latest")},
							Provider: "local",
							Regions:  []string{"north"},
						},
					},
				},
			},
		}

		imageConfig = image.NewImageConfiguration(config)
		mutator = NewMutator(imageConfig)
	})

	Describe("#Mutate", func() {
		It("should mutate all relevant container images", func() {
			cluster := &extensionscontroller.Cluster{
				Shoot: &gardencorev1beta1.Shoot{
					Spec: gardencorev1beta1.ShootSpec{
						Provider: gardencorev1beta1.Provider{
							Type: "local",
						},
						Region: "north",
					},
				},
			}

			ctx := context.WithValue(context.Background(), extensionswebhook.ClusterObjectContextKey{}, cluster)

			pod := &corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{Image: "another-init-image:latest"},
						{Image: "init-source-image:latest"},
					},
					Containers: []corev1.Container{
						{Image: "source-image:latest"},
						{Image: "another-image:latest"},
					},
				},
			}

			Expect(mutator.Mutate(ctx, pod, nil)).To(Succeed())
			Expect(pod.Spec.InitContainers).To(ConsistOf(
				corev1.Container{Image: "another-init-image:latest"},
				corev1.Container{Image: "init-target-image:latest"},
			))
			Expect(pod.Spec.Containers).To(ConsistOf(
				corev1.Container{Image: "target-image:latest"},
				corev1.Container{Image: "another-image:latest"},
			))
		})
	})
})
