// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"sync/atomic"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/utils/image"
)

const (
	// Type is the type of Extension resource.
	Type = "image-rewriter"
	// ControllerName is the name of the image rewriter controller.
	ControllerName = "image-rewriter-controller"
	// FinalizerSuffix is the finalizer suffix for the image rewriter controller.
	FinalizerSuffix = "image-rewriter"
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

// AddToManager adds the extension controller with the default Options to the given Controller Manager.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return extension.Add(mgr, extension.AddArgs{
		Actuator:          NewActuator(mgr.GetClient(), DefaultAddOptions.ShootWebhookConfig, image.NewImageConfiguration(&DefaultAddOptions.Config)),
		ControllerOptions: DefaultAddOptions.Controller,
		Name:              ControllerName,
		FinalizerSuffix:   FinalizerSuffix,
		Resync:            0,
		Predicates:        extension.DefaultPredicates(ctx, mgr, false),
		Type:              Type,
	})
}
