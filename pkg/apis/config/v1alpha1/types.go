// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains information about the registry service configuration.
type Configuration struct {
	metav1.TypeMeta `json:",inline"`

	// Overwrites configure the source and target images that should be replaced.
	// +optional
	Overwrites []ImageOverwrite `json:"overwrites,omitempty"`
}

// ImageOverwrite contains information about an image overwrite configuration.
type ImageOverwrite struct {
	// Source is the source image string to be replaced.
	Source Image `json:"source"`
	// Targets are the target images to replace the source with.
	Targets []TargetConfiguration `json:"targets"`
}

// TargetConfiguration contains information about the target image configuration.
type TargetConfiguration struct {
	Image `json:",inline"`
	// Provider is the name of the provider for which this target is applicable.
	Provider string `json:"provider"`
	// Regions are the regions where the target image is located.
	Regions []string `json:"regions"`
}

// Image contains information about an image.
type Image struct {
	// Image is the target image string to relace the source with.
	// +optional
	Image *string `json:"image,omitempty"`
	// Prefix is the prefix of the target image to relace the source with.
	// +optional
	Prefix *string `json:"prefix,omitempty"`
}
