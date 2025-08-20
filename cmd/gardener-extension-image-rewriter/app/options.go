// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"os"

	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	heartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/cmd"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/controller"
)

// ExtensionName is the name of the extension.
const ExtensionName = "extension-image-rewriter"

// Options holds configuration passed to the registry service controller.
type Options struct {
	extensionOptions   *cmd.ExtensionOptions
	restOptions        *controllercmd.RESTOptions
	managerOptions     *controllercmd.ManagerOptions
	controllerOptions  *controllercmd.ControllerOptions
	heartbeatOptions   *heartbeatcmd.Options
	controllerSwitches *controllercmd.SwitchOptions
	reconcileOptions   *controllercmd.ReconcilerOptions
	webhookOptions     *webhookcmd.AddToManagerOptions
	optionAggregator   controllercmd.OptionAggregator
}

// NewOptions creates a new Options instance.
func NewOptions() *Options {
	// options for the webhook server
	webhookServerOptions := &webhookcmd.ServerOptions{
		Namespace: os.Getenv("WEBHOOK_CONFIG_NAMESPACE"),
	}

	webhookSwitches := cmd.WebhookSwitchOptions()
	webhookOptions := webhookcmd.NewAddToManagerOptions(
		"image-rewriter",
		controller.ShootWebhooksResourceName,
		nil,
		webhookServerOptions,
		webhookSwitches,
	)

	options := &Options{
		restOptions: &controllercmd.RESTOptions{},
		managerOptions: &controllercmd.ManagerOptions{
			// These are default values.
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(ExtensionName),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
			WebhookServerPort:       443,
			WebhookCertDir:          "/tmp/gardener-extensions-cert",
			MetricsBindAddress:      ":8080",
			HealthBindAddress:       ":8081",
		},
		controllerOptions: &controllercmd.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		heartbeatOptions: &heartbeatcmd.Options{
			// This is a default value.
			ExtensionName:        ExtensionName,
			RenewIntervalSeconds: 30,
			Namespace:            os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		controllerSwitches: cmd.ControllerSwitches(),
		extensionOptions:   &cmd.ExtensionOptions{},
		reconcileOptions:   &controllercmd.ReconcilerOptions{},
		webhookOptions:     webhookOptions,
	}

	options.optionAggregator = controllercmd.NewOptionAggregator(
		options.extensionOptions,
		options.restOptions,
		options.managerOptions,
		options.controllerOptions,
		controllercmd.PrefixOption("heartbeat-", options.heartbeatOptions),
		options.controllerSwitches,
		options.reconcileOptions,
		options.webhookOptions,
	)

	return options
}
