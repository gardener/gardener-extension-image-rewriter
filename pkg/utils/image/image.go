// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"strings"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
)

// Configuration defines the interface for operating on image configurations.
type Configuration interface {
	// FindTargetImage returns the target image for a given source image, provider, and region.
	FindTargetImage(source string, provider string, region string) string
	// HasOverwrite checks if there is an overwrite for the given provider and region.
	HasOverwrite(provider string, region string) bool
}

type configuration struct {
	overwrites []overwrite
}

type overwrite struct {
	prefixed         bool
	source           string
	providerToTarget map[string]target
}

type target struct {
	regionToTarget map[string]string
}

// HasOverwrite checks if there is an overwrite for the given provider and region.
func (c *configuration) HasOverwrite(provider string, region string) bool {
	for _, overwrite := range c.overwrites {
		if target, exists := overwrite.providerToTarget[provider]; exists {
			if _, exists := target.regionToTarget[region]; exists {
				return true
			}
		}
	}
	return false
}

// FindTargetImage returns the target image for a given source image, provider, and region.
func (c *configuration) FindTargetImage(sourceImage string, provider string, region string) string {
	for _, overwrite := range c.overwrites {
		var imageSuffix string
		if overwrite.prefixed {
			if !strings.HasPrefix(sourceImage, overwrite.source) {
				continue
			}
			imageSuffix = strings.TrimPrefix(sourceImage, overwrite.source)
		} else {
			if overwrite.source != sourceImage {
				continue
			}
		}

		target, providerConfigured := overwrite.providerToTarget[provider]
		if !providerConfigured {
			continue
		}

		targetImage, imageConfigured := target.regionToTarget[region]
		if !imageConfigured {
			continue
		}

		if overwrite.prefixed {
			return targetImage + imageSuffix
		}
		return targetImage
	}

	return ""
}

// NewImageConfiguration creates a new image configuration implementation.
func NewImageConfiguration(config *v1alpha1.Configuration) Configuration {
	overwrites := make([]overwrite, 0, len(config.Overwrites))
	for _, o := range config.Overwrites {
		providerToTarget := make(map[string]target)
		for _, t := range o.Targets {
			if _, exists := providerToTarget[t.Provider]; !exists {
				providerToTarget[t.Provider] = target{
					regionToTarget: make(map[string]string),
				}
			}

			for _, region := range t.Regions {
				providerToTarget[t.Provider].regionToTarget[region] = prefixOrImage(t.Image)
			}
		}

		overwrites = append(overwrites, overwrite{
			prefixed:         o.Source.Prefix != nil,
			source:           prefixOrImage(o.Source),
			providerToTarget: providerToTarget,
		})
	}

	return &configuration{
		overwrites: overwrites,
	}
}

func prefixOrImage(image v1alpha1.Image) string {
	if image.Prefix != nil {
		return *image.Prefix
	}
	if image.Image != nil {
		return *image.Image
	}
	return ""
}
