// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package pod

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/image"
)

type mutator struct {
	config image.Configuration
}

var _ extensionswebhook.WantsClusterObject = (*mutator)(nil)

// NewMutator creates a new Mutator instance.
func NewMutator(config image.Configuration) extensionswebhook.Mutator {
	return &mutator{
		config: config,
	}
}

// Mutate mutates the given Pod object by replacing the images of its containers if a replacement is defined.
func (m *mutator) Mutate(ctx context.Context, new, _ client.Object) error {
	log := logf.FromContext(ctx)

	// Get Cluster Object from context
	clusterValue := ctx.Value(extensionswebhook.ClusterObjectContextKey{})
	if clusterValue == nil {
		return fmt.Errorf("cluster not found in context")
	}

	cluster, ok := clusterValue.(*extensionscontroller.Cluster)
	if !ok {
		return fmt.Errorf("expected object to be of type *extensionscontroller.Cluster, got %T", new)
	}

	pod, ok := new.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected new object to be of type *corev1.Pod, got %T", new)
	}

	for i, container := range pod.Spec.InitContainers {
		if image := m.config.FindTargetImage(container.Image, cluster.Shoot.Spec.Provider.Type, cluster.Shoot.Spec.Region); image != "" {
			log.V(2).Info("Replacing container image", "oldImage", container.Image, "newImage", image)
			pod.Spec.InitContainers[i].Image = image
		}
	}

	for i, container := range pod.Spec.Containers {
		if image := m.config.FindTargetImage(container.Image, cluster.Shoot.Spec.Provider.Type, cluster.Shoot.Spec.Region); image != "" {
			log.V(2).Info("Replacing container image", "oldImage", container.Image, "newImage", image)
			pod.Spec.Containers[i].Image = image
		}
	}

	return nil
}

func (m *mutator) WantsClusterObject() bool {
	return true
}
