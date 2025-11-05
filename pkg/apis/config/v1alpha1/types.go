// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
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

	// ContainerdConfiguration contains the containerd configuration for the image rewriter.
	// +optional
	Containerd []ContainerdConfiguration `json:"containerd,omitempty"`
	// Overwrites configure the source and target images that should be replaced.
	// +optional
	Overwrites []ImageOverwrite `json:"overwrites,omitempty"`
}

// ContainerdConfiguration contains information about a containerd upstream configuration.
type ContainerdConfiguration struct {
	// Upstream is the upstream name of the registry.
	Upstream string `json:"upstream"`
	// Server is the URL of the upstream registry.
	Server string `json:"server"`
	// Hosts are the containerd hosts separated by provider and regions.
	Hosts []ContainerdHostConfig `json:"hosts"`
}

// ContainerdHostConfig contains information about a containerd host configuration.
type ContainerdHostConfig struct {
	URL string `json:"url"`
	// Provider is the name of the provider for which this target is applicable.
	Provider string `json:"provider"`
	// Regions are the regions where the target image is located. If not specified, any shoot region will match this host config.
	// +optional
	Regions []string `json:"regions,omitempty"`
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
