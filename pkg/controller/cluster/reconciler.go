// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"context"
	"fmt"
	"sync/atomic"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/extensions/pkg/webhook/shoot"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	"github.com/labstack/gommon/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/utils/image"
)

type reconciler struct {
	client client.Client

	shootWebhookConfig *atomic.Value
	config             image.Configuration
}

// ShootWebhooksResourceName is the name of the managed resource for the Shoot webhooks.
const ShootWebhooksResourceName = "extension-image-rewriter-shoot-webhooks"

func (r *reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	cluster, err := extensionscontroller.GetCluster(ctx, r.client, req.Name)
	if client.IgnoreNotFound(err) != nil {
		return reconcile.Result{}, err
	}

	log := logf.FromContext(ctx)

	if cluster == nil || cluster.Shoot.DeletionTimestamp != nil {
		return r.delete(ctx, req.Name)
	}

	return r.reconcile(ctx, log, cluster)
}

func (r *reconciler) reconcile(ctx context.Context, log logr.Logger, cluster *extensionscontroller.Cluster) (reconcile.Result, error) {
	if !r.config.HasOverwrite(cluster.Shoot.Spec.Provider.Type, cluster.Shoot.Spec.Region) {
		log.Info("No overwrite configuration found for shoot provider and region")
		return r.delete(ctx, cluster.ObjectMeta.Name)
	}

	if err := r.reconcileShootWebhookConfig(ctx, cluster); err != nil {
		log.Error(err, "Failed to reconcile Shoot webhook configuration")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) reconcileShootWebhookConfig(ctx context.Context, cluster *extensionscontroller.Cluster) error {
	value := r.shootWebhookConfig.Load()
	webhookConfig, ok := value.(*webhook.Configs)
	if !ok {
		return fmt.Errorf("expected *webhook.Configs, got %T", value)
	}

	if err := shoot.ReconcileWebhookConfig(ctx, r.client, cluster.ObjectMeta.Name, ShootWebhooksResourceName, *webhookConfig, cluster, true); err != nil {
		return fmt.Errorf("could not reconcile shoot webhooks: %w", err)
	}

	return nil
}

func (r *reconciler) delete(ctx context.Context, clusterName string) (reconcile.Result, error) {
	log.Info("Deleting Shoot webhook configuration")
	return reconcile.Result{}, managedresources.DeleteForShoot(ctx, r.client, clusterName, ShootWebhooksResourceName)
}
