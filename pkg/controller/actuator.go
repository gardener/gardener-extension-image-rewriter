// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"sync/atomic"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	"github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/extensions/pkg/webhook/shoot"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/utils/image"
)

type actuator struct {
	client client.Client

	shootWebhookConfig *atomic.Value
	config             image.Configuration
}

// NewActuator returns an actuator responsible for registry-cache Extension resources.
func NewActuator(client client.Client, shootWebhookConfig *atomic.Value, config image.Configuration) extension.Actuator {
	return &actuator{
		client:             client,
		shootWebhookConfig: shootWebhookConfig,
		config:             config,
	}
}

// ShootWebhooksResourceName is the name of the managed resource for the Shoot webhooks.
const ShootWebhooksResourceName = "extension-image-rewriter-shoot-webhooks"

// Reconcile reconciles the Extension resource. It creates or deletes the shoot webhook configuration, depending on whether an overwrite configuration exists for the shoot's provider and region.
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, e *extensionsv1alpha1.Extension) error {
	cluster, err := extensionscontroller.GetCluster(ctx, a.client, e.Namespace)
	if client.IgnoreNotFound(err) != nil {
		return err
	}

	if !a.config.HasOverwrite(cluster.Shoot.Spec.Provider.Type, cluster.Shoot.Spec.Region) {
		log.Info("No overwrite configuration found for shoot provider and region")
		return a.Delete(ctx, log, e)
	}

	return a.reconcileShootWebhookConfig(ctx, cluster)
}

func (a *actuator) reconcileShootWebhookConfig(ctx context.Context, cluster *extensionscontroller.Cluster) error {
	value := a.shootWebhookConfig.Load()
	webhookConfig, ok := value.(*webhook.Configs)
	if !ok {
		return fmt.Errorf("expected *webhook.Configs, got %T", value)
	}

	if err := shoot.ReconcileWebhookConfig(ctx, a.client, cluster.ObjectMeta.Name, ShootWebhooksResourceName, *webhookConfig, cluster, true); err != nil {
		return fmt.Errorf("could not reconcile shoot webhooks: %w", err)
	}

	return nil
}

// Delete deletes the Extension resource.
func (a *actuator) Delete(ctx context.Context, log logr.Logger, e *extensionsv1alpha1.Extension) error {
	log.Info("Deleting Shoot webhook configuration")
	return managedresources.DeleteForShoot(ctx, a.client, e.Namespace, ShootWebhooksResourceName)
}

// ForceDelete forcefully deletes the Extension resource.
func (a *actuator) ForceDelete(ctx context.Context, log logr.Logger, e *extensionsv1alpha1.Extension) error {
	return a.Delete(ctx, log, e)
}

// Restore restores the Extension resource. Calling this method is a no-op for this actuator.
func (a *actuator) Restore(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.Extension) error {
	return nil
}

// Migrate migrates the Extension resource. Calling this method is a no-op for this actuator.
func (a *actuator) Migrate(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.Extension) error {
	return nil
}
