// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cluster_test

import (
	"context"
	"sync/atomic"

	"github.com/gardener/gardener/extensions/pkg/webhook"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	. "github.com/gardener/gardener/pkg/utils/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	clustercontroller "github.com/gardener/gardener-extension-image-rewriter/pkg/controller/cluster"
)

var _ = Describe("Cluster controller test", func() {
	var cluster *extensionsv1alpha1.Cluster

	BeforeEach(func() {
		cluster = &extensionsv1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: testRunID,
				Labels: map[string]string{
					testID: testRunID,
				},
			},
			Spec: extensionsv1alpha1.ClusterSpec{
				CloudProfile: runtime.RawExtension{
					Raw: []byte(`{}`),
				},
				Seed: runtime.RawExtension{
					Raw: []byte(`{}`),
				},
				Shoot: runtime.RawExtension{
					Raw: []byte(`{
"apiVersion": "core.gardener.cloud/v1beta1",
"kind": "Shoot",
"spec": {
"provider": {
  "type": "local-foo"
},
"region": "north"
}
}`),
				},
			},
		}
	})

	Describe("Create the webhook configuration", Ordered, func() {
		It("should add the configuration and controller to the manager", func() {
			clustercontroller.DefaultAddOptions.Config = v1alpha1.Configuration{
				Overwrites: []v1alpha1.ImageOverwrite{
					{
						Source: v1alpha1.Image{Prefix: ptr.To("gardener.cloud/gardener-project")},
						Targets: []v1alpha1.TargetConfiguration{
							{
								Image:    v1alpha1.Image{Prefix: ptr.To("registry.central.local/gardener-project")},
								Provider: "local",
								Region:   "central",
							},
						},
					},
				},
			}

			clustercontroller.DefaultAddOptions.ShootWebhookConfig = &atomic.Value{}
			clustercontroller.DefaultAddOptions.ShootWebhookConfig.Store(&webhook.Configs{
				MutatingWebhookConfig: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: "extension-image-rewriter-shoot-webhooks",
					},
				},
			})

			Expect(clustercontroller.AddToManager(ctx, mgr)).To(Succeed())
		})

		It("should do nothing because no overwrite configuration is found", func() {
			Expect(mgrClient.Create(ctx, cluster)).To(Succeed())

			DeferCleanup(func() {
				Expect(mgrClient.Delete(ctx, cluster)).To(Or(Succeed(), BeNotFoundError()))

				Eventually(func() error {
					return mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), &extensionsv1alpha1.Cluster{})
				}).To(BeNotFoundError())
			})

			Consistently(func() int {
				managedResources := &resourcesv1alpha1.ManagedResourceList{}
				Expect(mgrClient.List(ctx, managedResources, client.InNamespace(testRunID))).To(Succeed())
				return len(managedResources.Items)
			}).To(Equal(0), "No managed resources should be created for the cluster with no overwrite configuration")
		})

		It("should add the webhook configuration to the cluster", func() {
			cluster.Spec.Shoot = runtime.RawExtension{
				Raw: []byte(`{
"apiVersion": "core.gardener.cloud/v1beta1",
"kind": "Shoot",
"spec": {
"provider": {
  "type": "local"
},
"region": "central"
}
}`)}

			Expect(mgrClient.Create(ctx, cluster)).To(Succeed())

			verifyWebhookConfig(ctx, mgrClient, testRunID, true)
		})

		It("should remove the webhook configuration from the cluster when no overwrite is found", func() {
			existingCluster := &extensionsv1alpha1.Cluster{}
			Expect(mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), existingCluster)).To(Succeed())

			existingCluster.Spec.Shoot = runtime.RawExtension{
				Raw: []byte(`{
"apiVersion": "core.gardener.cloud/v1beta1",
"kind": "Shoot",
"spec": {
"provider": {
  "type": "local"
},
"region": "north"
}
}`)}

			Expect(mgrClient.Update(ctx, existingCluster)).To(Succeed())

			verifyWebhookConfig(ctx, mgrClient, testRunID, false)
		})

		It("should remove the webhook when cluster is deleted", func() {
			existingCluster := &extensionsv1alpha1.Cluster{}
			Expect(mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), existingCluster)).To(Succeed())

			existingCluster.Spec.Shoot = runtime.RawExtension{
				Raw: []byte(`{
"apiVersion": "core.gardener.cloud/v1beta1",
"kind": "Shoot",
"spec": {
"provider": {
  "type": "local"
},
"region": "central"
}
}`)}
			Expect(mgrClient.Update(ctx, existingCluster)).To(Succeed())

			verifyWebhookConfig(ctx, mgrClient, testRunID, true)

			Expect(mgrClient.Delete(ctx, existingCluster)).To(Succeed())

			verifyWebhookConfig(ctx, mgrClient, testRunID, false)
		})
	})
})

func verifyWebhookConfig(ctx context.Context, cl client.Client, namespace string, expectedAvailable bool) {
	GinkgoHelper()

	expectedElements := 0
	description := "No managed resources should be created for the cluster with no overwrite configuration"

	if expectedAvailable {
		expectedElements = 1
		description = "Managed resource should be created for the cluster with overwrite configuration"
	}

	Eventually(func() int {
		managedResources := &resourcesv1alpha1.ManagedResourceList{}
		Expect(cl.List(ctx, managedResources, client.InNamespace(namespace))).To(Succeed())
		return len(managedResources.Items)
	}).To(Equal(expectedElements), description)
}
