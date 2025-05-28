// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package pod

import (
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/extensions/pkg/webhook/shoot"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/image"
)

const (
	// Name is the name of the webhook.
	Name = "pod-image-rewriter"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the AWS shoot webhook to the manager.
type AddOptions struct {
	Config v1alpha1.Configuration
}

// AddToManager creates a webhook with the DefaultAddOptions.
func AddToManager(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	log.Log.Info("Adding webhook to manager")

	return shoot.New(mgr, shoot.Args{
		Types: []extensionswebhook.Type{
			{Obj: &corev1.Pod{}},
		},
		Mutator:       NewMutator(image.NewImageConfiguration(&DefaultAddOptions.Config)),
		FailurePolicy: ptr.To(admissionregistrationv1.Ignore),
	})
}
