// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"os"

	"github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	webhookcmd "github.com/gardener/gardener/extensions/pkg/webhook/cmd"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/validation"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/controller"
	containerdwebhook "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/operatingsystemconfig/containerd"
	imagewebhook "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/operatingsystemconfig/image"
	podwebhook "github.com/gardener/gardener-extension-image-rewriter/pkg/webhook/pod"
)

var (
	scheme  *runtime.Scheme
	decoder runtime.Decoder
)

func init() {
	scheme = runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
}

// ExtensionOptions holds options related to the image rewriter.
type ExtensionOptions struct {
	ConfigLocation string
	config         *ExtensionConfig
}

// AddFlags implements Flagger.AddFlags.
func (o *ExtensionOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigLocation, "config", "", "Path to image rewriter configuration")
}

// Complete implements Completer.Complete.
func (o *ExtensionOptions) Complete() error {
	if o.ConfigLocation == "" {
		return errors.New("config location is not set")
	}
	data, err := os.ReadFile(o.ConfigLocation)
	if err != nil {
		return err
	}

	config := v1alpha1.Configuration{}
	if err := runtime.DecodeInto(decoder, data, &config); err != nil {
		return err
	}

	if errs := validation.ValidateConfiguration(&config); len(errs) > 0 {
		return errs.ToAggregate()
	}

	o.config = &ExtensionConfig{
		config: config,
	}

	return nil
}

// Completed returns the decoded ExtensionConfiguration instance. Only call this if `Complete` was successful.
func (o *ExtensionOptions) Completed() *ExtensionConfig {
	return o.config
}

// ExtensionConfig contains configuration information about the image rewriter.
type ExtensionConfig struct {
	config v1alpha1.Configuration
}

// Apply applies the ExtensionOptions to the passed ControllerOptions instance.
func (c *ExtensionConfig) Apply(config *v1alpha1.Configuration) {
	*config = c.config
}

// ControllerSwitches are the cmd.SwitchOptions for the provider controllers.
func ControllerSwitches() *cmd.SwitchOptions {
	return cmd.NewSwitchOptions(
		cmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
		cmd.Switch(controller.ControllerName, controller.AddToManager),
	)
}

// WebhookSwitchOptions are the webhookcmd.SwitchOptions for the image rewriter webhook.
func WebhookSwitchOptions() *webhookcmd.SwitchOptions {
	return webhookcmd.NewSwitchOptions(
		webhookcmd.Switch(podwebhook.Name, podwebhook.AddToManager),
		webhookcmd.Switch(imagewebhook.Name, imagewebhook.AddToManager),
		webhookcmd.Switch(containerdwebhook.Name, containerdwebhook.AddToManager),
	)
}
