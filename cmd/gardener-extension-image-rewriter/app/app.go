// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller"
	heartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	"github.com/gardener/gardener/extensions/pkg/util"
	gardenerhealthz "github.com/gardener/gardener/pkg/healthz"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/component-base/version"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	clustercontroller "github.com/gardener/gardener-extension-image-rewriter/pkg/controller/cluster"
	containerdwebhook "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/operatingsystemconfig/containerd"
	imagewebhook "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/operatingsystemconfig/image"
	podwebhook "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/pod"
)

var log = logf.Log.WithName("gardener-extension-image-rewriter")

// NewServiceControllerCommand creates a new command that is used to start the registry service controller.
func NewServiceControllerCommand() *cobra.Command {
	options := NewOptions()

	cmd := &cobra.Command{
		Use:           "image-rewriter",
		Short:         "Image rewriter rewrites image paths of pods within a shoot.",
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, _ []string) error {
			verflag.PrintAndExitIfRequested()

			log.Info("Starting image-rewriter", "version", version.Get())

			if err := options.optionAggregator.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			if err := options.heartbeatOptions.Validate(); err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return options.run(cmd.Context())
		},
	}

	verflag.AddFlags(cmd.Flags())
	options.optionAggregator.AddFlags(cmd.Flags())

	return cmd
}

func (o *Options) run(ctx context.Context) error {
	util.ApplyClientConnectionConfigurationToRESTConfig(&componentbaseconfigv1alpha1.ClientConnectionConfiguration{
		QPS:   100.0,
		Burst: 130,
	}, o.restOptions.Completed().Config)

	mgrOpts := o.managerOptions.Completed().Options()

	mgrOpts.Client = client.Options{
		Cache: &client.CacheOptions{
			DisableFor: []client.Object{
				&corev1.Secret{}, // applied for ManagedResources
			},
		},
	}

	mgr, err := manager.New(o.restOptions.Completed().Config, mgrOpts)
	if err != nil {
		return fmt.Errorf("could not instantiate controller-manager: %w", err)
	}

	scheme := mgr.GetScheme()
	if err := controller.AddToScheme(scheme); err != nil {
		return fmt.Errorf("could not update manager scheme: %w", err)
	}

	o.heartbeatOptions.Completed().Apply(&heartbeatcontroller.DefaultAddOptions)
	o.controllerOptions.Completed().Apply(&clustercontroller.DefaultAddOptions.Controller)
	o.extensionOptions.Completed().Apply(&clustercontroller.DefaultAddOptions.Config)
	o.extensionOptions.Completed().Apply(&podwebhook.DefaultAddOptions.Config)
	o.extensionOptions.Completed().Apply(&imagewebhook.DefaultAddOptions.Config)
	o.extensionOptions.Completed().Apply(&containerdwebhook.DefaultAddOptions.Config)
	shootWebhookConfig, err := o.webhookOptions.Completed().AddToManager(ctx, mgr, nil, false)
	if err != nil {
		return fmt.Errorf("could not add the mutating webhook to manager: %w", err)
	}
	clustercontroller.DefaultAddOptions.ShootWebhookConfig = shootWebhookConfig

	if err := o.controllerSwitches.Completed().AddToManager(ctx, mgr); err != nil {
		return fmt.Errorf("could not add controllers to manager: %w", err)
	}

	if err := mgr.AddReadyzCheck("informer-sync", gardenerhealthz.NewCacheSyncHealthz(mgr.GetCache())); err != nil {
		return fmt.Errorf("could not add ready check for informers: %w", err)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("could not add health check to manager: %w", err)
	}

	if err := mgr.AddReadyzCheck("webhook-server", mgr.GetWebhookServer().StartedChecker()); err != nil {
		return fmt.Errorf("could not add ready check for webhook server to manager: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}

	return nil
}
