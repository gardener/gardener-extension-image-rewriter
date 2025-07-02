// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package image

import (
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
)

const (
	// Name is the name of the webhook.
	Name = "osc-image-rewriter"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the AWS shoot webhook to the manager.
type AddOptions struct {
	Config v1alpha1.Configuration
}

// AddToManager creates a webhook and adds it to the manager.
func AddToManager(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	logger := log.Log.WithValues("webhook", Name)
	logger.Info("Adding webhook to manager")

	// Create handler
	types := []extensionswebhook.Type{
		{Obj: &extensionsv1alpha1.OperatingSystemConfig{}},
	}

	handler, err := extensionswebhook.NewBuilder(mgr, logger).WithMutator(NewMutator(mgr.GetClient(), &DefaultAddOptions.Config), types...).Build()
	if err != nil {
		return nil, err
	}

	return &extensionswebhook.Webhook{
		Name:    Name,
		Types:   types,
		Target:  extensionswebhook.TargetSeed,
		Path:    Name,
		Webhook: &admission.Webhook{Handler: handler, RecoverPanic: ptr.To(true)},
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				v1beta1constants.GardenRole: v1beta1constants.GardenRoleShoot,
			},
		},
	}, nil
}
