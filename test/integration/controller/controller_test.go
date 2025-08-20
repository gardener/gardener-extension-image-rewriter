// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller_test

import (
	"context"
	"sync/atomic"

	"github.com/gardener/gardener/extensions/pkg/webhook"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
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
	"github.com/gardener/gardener-extension-image-rewriter/pkg/controller"
)

var _ = Describe("Cluster controller test", func() {
	var (
		cluster   *extensionsv1alpha1.Cluster
		extension *extensionsv1alpha1.Extension

		rawShootCentral, rawShootNorth []byte
	)

	BeforeEach(func() {
		rawShootCentral = []byte(`{
"apiVersion": "core.gardener.cloud/v1beta1",
"kind": "Shoot",
"spec": {
"provider": {
  "type": "local"
},
"region": "central"
}
}`)
		rawShootNorth = []byte(`{
"apiVersion": "core.gardener.cloud/v1beta1",
"kind": "Shoot",
"spec": {
"provider": {
  "type": "local"
},
"region": "north"
}
}`)

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
			},
		}

		extension = &extensionsv1alpha1.Extension{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testRunID,
				Namespace: testNamespace.Name,
				Labels: map[string]string{
					testID: testRunID,
				},
			},
			Spec: extensionsv1alpha1.ExtensionSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type: "image-rewriter",
				},
			},
		}
	})

	Describe("Create the webhook configuration", Ordered, func() {
		It("should add the configuration and controller to the manager", func() {
			controller.DefaultAddOptions.Config = v1alpha1.Configuration{
				Overwrites: []v1alpha1.ImageOverwrite{
					{
						Source: v1alpha1.Image{Prefix: ptr.To("gardener.cloud/gardener-project")},
						Targets: []v1alpha1.TargetConfiguration{
							{
								Image:    v1alpha1.Image{Prefix: ptr.To("registry.central.local/gardener-project")},
								Provider: "local",
								Regions:  []string{"central"},
							},
						},
					},
				},
			}

			controller.DefaultAddOptions.ShootWebhookConfig = &atomic.Value{}
			controller.DefaultAddOptions.ShootWebhookConfig.Store(&webhook.Configs{
				MutatingWebhookConfig: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: "extension-image-rewriter-shoot-webhooks",
					},
				},
			})

			Expect(controller.AddToManager(ctx, mgr)).To(Succeed())
		})

		It("should do nothing because no overwrite configuration is found", func() {
			cluster.Spec.Shoot.Raw = rawShootNorth
			Expect(mgrClient.Create(ctx, cluster)).To(Succeed())
			Expect(mgrClient.Create(ctx, extension)).To(Succeed())

			DeferCleanup(func() {
				Expect(mgrClient.Delete(ctx, cluster)).To(Or(Succeed(), BeNotFoundError()))
				Eventually(func() error {
					return mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), &extensionsv1alpha1.Cluster{})
				}).To(BeNotFoundError())

				Expect(mgrClient.Delete(ctx, extension)).To(Or(Succeed(), BeNotFoundError()))
				Eventually(func() error {
					return mgrClient.Get(ctx, client.ObjectKeyFromObject(extension), &extensionsv1alpha1.Extension{})
				}).To(BeNotFoundError())
			})

			triggerExtensionReconciliation(extension)
			waitForExtensionReconciliation(extension)

			managedResources := &resourcesv1alpha1.ManagedResourceList{}
			Expect(mgrClient.List(ctx, managedResources, client.InNamespace(testRunID))).To(Succeed())
			Expect(managedResources.Items).To(BeEmpty(), "No managed resources should be created for the cluster with no overwrite configuration")
		})

		It("should add the webhook configuration to the cluster", func() {
			cluster.Spec.Shoot.Raw = rawShootCentral
			Expect(mgrClient.Create(ctx, cluster)).To(Succeed())
			Expect(mgrClient.Create(ctx, extension)).To(Succeed())
			waitForExtensionReconciliation(extension)
			verifyWebhookConfig(ctx, mgrClient, testRunID, true)
		})

		It("should remove the webhook configuration from the cluster when no overwrite is found", func() {
			Expect(mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), cluster)).To(Succeed())
			cluster.Spec.Shoot.Raw = rawShootNorth
			Expect(mgrClient.Update(ctx, cluster)).To(Succeed())
			Eventually(func(g Gomega) {
				existingCluster := &extensionsv1alpha1.Cluster{}
				g.Expect(mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), existingCluster)).To(Succeed())
				g.Expect(string(existingCluster.Spec.Shoot.Raw)).To(ContainSubstring(`"region":"north"`))
			}).Should(Succeed())

			triggerExtensionReconciliation(extension)
			waitForExtensionReconciliation(extension)
			verifyWebhookConfig(ctx, mgrClient, testRunID, false)
		})

		It("should remove the webhook when extension is deleted", func() {
			Expect(mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), cluster)).To(Succeed())
			cluster.Spec.Shoot.Raw = rawShootCentral
			Expect(mgrClient.Update(ctx, cluster)).To(Succeed())
			Eventually(func(g Gomega) {
				existingCluster := &extensionsv1alpha1.Cluster{}
				g.Expect(mgrClient.Get(ctx, client.ObjectKeyFromObject(cluster), existingCluster)).To(Succeed())
				g.Expect(string(existingCluster.Spec.Shoot.Raw)).To(ContainSubstring(`"region":"central"`))
			}).Should(Succeed())

			triggerExtensionReconciliation(extension)
			waitForExtensionReconciliation(extension)
			verifyWebhookConfig(ctx, mgrClient, testRunID, true)

			Expect(mgrClient.Delete(ctx, extension)).To(Succeed())
			Eventually(func() error {
				return mgrClient.Get(ctx, client.ObjectKeyFromObject(extension), &extensionsv1alpha1.Extension{})
			}).To(BeNotFoundError())
			verifyWebhookConfig(ctx, mgrClient, testRunID, false)
		})
	})
})

func triggerExtensionReconciliation(extension *extensionsv1alpha1.Extension) {
	GinkgoHelper()

	patch := client.MergeFrom(extension.DeepCopy())
	metav1.SetMetaDataAnnotation(&extension.ObjectMeta, "gardener.cloud/operation", "reconcile")
	Expect(mgrClient.Patch(ctx, extension, patch)).To(Succeed())
}

func waitForExtensionReconciliation(extension *extensionsv1alpha1.Extension) {
	GinkgoHelper()

	Eventually(func(g Gomega) {
		g.Expect(mgrClient.Get(ctx, client.ObjectKeyFromObject(extension), extension)).To(Succeed())
		g.Expect(extension.Generation).To(Equal(extension.Status.ObservedGeneration))
		g.Expect(extension.Annotations).NotTo(HaveKey("gardener.cloud/operation"))
		g.Expect(extension.Status.LastOperation).NotTo(BeNil())
		g.Expect(extension.Status.LastOperation.State).To(Equal(gardencorev1beta1.LastOperationStateSucceeded))
	}).Should(Succeed())
}

func verifyWebhookConfig(ctx context.Context, cl client.Client, namespace string, expectedAvailable bool) {
	GinkgoHelper()

	expectedElements := 0
	description := "No managed resources should be created for the cluster with no overwrite configuration"

	if expectedAvailable {
		expectedElements = 1
		description = "Managed resource should be created for the cluster with overwrite configuration"
	}

	Eventually(func(g Gomega) {
		managedResources := &resourcesv1alpha1.ManagedResourceList{}
		g.Expect(cl.List(ctx, managedResources, client.InNamespace(namespace))).To(Succeed())
		g.Expect(managedResources.Items).To(HaveLen(expectedElements), description)
	}).Should(Succeed())
}
