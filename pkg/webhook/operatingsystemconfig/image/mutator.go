// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"
	"regexp"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	extensionsv1alpha1helper "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1/helper"
	gardenerutils "github.com/gardener/gardener/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-extension-image-rewriter/pkg/image"
)

type mutator struct {
	client client.Client
	config image.Configuration
}

// Regex matches:
// 1. prefix@sha256:checksum
// 2. prefix:version
// 3. prefix:version@sha256:checksum
var ociImagePattern = regexp.MustCompile(`\b[\w\-\.\/]+:(?:[\w\.\-]+@sha256:[a-fA-F0-9]{64}|[\w\.\-]+)|[\w\-\.\/]+@sha256:[a-fA-F0-9]{64}\b`)

func (m *mutator) Mutate(ctx context.Context, new, _ client.Object) error {
	log := logf.FromContext(ctx)

	cluster, err := extensionscontroller.GetCluster(ctx, m.client, new.GetNamespace())
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	osc, ok := new.(*extensionsv1alpha1.OperatingSystemConfig)
	if !ok {
		return fmt.Errorf("expected new object to be of type *extensionsv1alpha1.OperatingSystemConfig, got %T", new)
	}

	var (
		shootProvider = cluster.Shoot.Spec.Provider.Type
		shootRegion   = cluster.Shoot.Spec.Region
	)

	switch osc.Spec.Purpose {
	case extensionsv1alpha1.OperatingSystemConfigPurposeReconcile:
		for i, file := range osc.Spec.Files {
			if file.Content.ImageRef != nil {
				if newImage := m.config.FindTargetImage(file.Content.ImageRef.Image, shootProvider, shootRegion); newImage != "" {
					log.V(2).Info("Replacing image in OperatingSystemConfig file", "oldImage", file.Content.ImageRef.Image, "newImage", newImage)
					osc.Spec.Files[i].Content.ImageRef.Image = newImage
				}
			}
		}

		if extensionsv1alpha1helper.HasContainerdConfiguration(osc.Spec.CRIConfig) {
			if newImage := m.config.FindTargetImage(osc.Spec.CRIConfig.Containerd.SandboxImage, shootProvider, shootRegion); newImage != "" {
				log.V(2).Info("Replacing sandbox image in OperatingSystemConfig file", "oldImage", osc.Spec.CRIConfig.Containerd.SandboxImage, "newImage", newImage)
				osc.Spec.CRIConfig.Containerd.SandboxImage = newImage
			}
		}

	case extensionsv1alpha1.OperatingSystemConfigPurposeProvision:
		for i, file := range osc.Spec.Files {
			if inlineContent := file.Content.Inline; inlineContent != nil {
				data, err := readData(inlineContent)
				if err != nil {
					return fmt.Errorf("failed to read file content: %w", err)
				}

				var updated bool
				data = ociImagePattern.ReplaceAllStringFunc(data, func(match string) string {
					if newImage := m.config.FindTargetImage(match, shootProvider, shootRegion); newImage != "" {
						log.V(2).Info("Replacing image in OperatingSystemConfig file", "oldImage", match, "newImage", newImage)
						updated = true
						return newImage
					}
					return match
				})

				if updated {
					writeData(osc.Spec.Files[i].Content.Inline, data)
				}
			}
		}
	}

	return nil
}

func readData(fileContent *extensionsv1alpha1.FileContentInline) (string, error) {
	if fileContent.Encoding == string(extensionsv1alpha1.B64FileCodecID) {
		decodedData, err := gardenerutils.DecodeBase64(fileContent.Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 content: %w", err)
		}
		return string(decodedData), nil
	}
	return fileContent.Data, nil
}

func writeData(fileContent *extensionsv1alpha1.FileContentInline, data string) {
	if fileContent.Encoding == string(extensionsv1alpha1.B64FileCodecID) {
		fileContent.Data = gardenerutils.EncodeBase64([]byte(data))
	} else {
		fileContent.Data = data
	}
}

// NewMutator creates a new Mutator instance.
func NewMutator(client client.Client, config *v1alpha1.Configuration) extensionswebhook.Mutator {
	return &mutator{
		client: client,
		config: image.NewImageConfiguration(config),
	}
}
