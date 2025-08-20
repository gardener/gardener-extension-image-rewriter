// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"sync/atomic"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/utils/image"
)

const (
	// ControllerName is the name of the image rewriter controller.
	ControllerName = "image-rewriter-cluster-controller"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the controller to the manager.
type AddOptions struct {
	// Controller contains options for the controller.
	Controller controller.Options
	// Config is the configuration for the image rewriter extension.
	Config v1alpha1.Configuration
	// ShootWebhookConfig holds the current Shoot webhook configuration.
	ShootWebhookConfig *atomic.Value
}

// AddToManager adds a controller with the default Options to the given Controller Manager.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	DefaultAddOptions.Controller.Reconciler = &reconciler{
		client:             mgr.GetClient(),
		config:             image.NewImageConfiguration(&DefaultAddOptions.Config),
		shootWebhookConfig: DefaultAddOptions.ShootWebhookConfig,
	}

	ctrl, err := controller.New(ControllerName, mgr, DefaultAddOptions.Controller)
	if err != nil {
		return err
	}

	return ctrl.Watch(
		source.Kind[client.Object](mgr.GetCache(),
			&extensionsv1alpha1.Cluster{},
			&handler.EnqueueRequestForObject{},
			predicate.NewPredicateFuncs(func(object client.Object) bool {
				return mgr.GetClient().Get(ctx, client.ObjectKey{Name: object.GetName()}, &corev1.Namespace{}) == nil
			}),
		),
	)
}
