// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package image_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	. "github.com/gardener/gardener-extension-image-rewriter/pkg/image"
)

var _ = Describe("Image", func() {
	var (
		image string

		imageConfig Configuration
		config      *v1alpha1.Configuration
	)

	BeforeEach(func() {
		image = "registry.example.com/image:latest"

		imageConfig = nil
		config = &v1alpha1.Configuration{
			Overwrites: []v1alpha1.ImageOverwrite{
				{
					Source: v1alpha1.Image{
						Image: ptr.To(image),
					},
					Targets: []v1alpha1.TargetConfiguration{
						{
							Image:    v1alpha1.Image{Image: imageReplacement("west")},
							Provider: "local",
							Region:   "west",
						},
						{
							Image:    v1alpha1.Image{Image: imageReplacement("east")},
							Provider: "local",
							Region:   "east",
						},
					},
				},
			},
		}
	})

	Describe("#FindTargetImage", func() {
		var prefixSource string

		BeforeEach(func() {
			prefixSource = image[:strings.Index(image, "/")]

			// Config with prefix source and target
			config.Overwrites = append(config.Overwrites, v1alpha1.ImageOverwrite{
				Source: v1alpha1.Image{
					Prefix: &prefixSource,
				},
				Targets: []v1alpha1.TargetConfiguration{
					{
						Image:    v1alpha1.Image{Prefix: imageReplacementPrefix("west")},
						Provider: "local2",
						Region:   "west",
					},
					{
						Image:    v1alpha1.Image{Prefix: imageReplacementPrefix("east")},
						Provider: "local2",
						Region:   "east",
					},
				},
			})

			imageConfig = NewImageConfiguration(config)
		})

		It("should find the target image with prefix", func() {
			expectedTargetImageWest := ptr.Deref(imageReplacementPrefix("west"), "") + "/foo:bar"
			Expect(imageConfig.FindTargetImage(prefixSource+"/foo:bar", "local2", "west")).To(Equal(expectedTargetImageWest))

			expectedTargetImageEast := ptr.Deref(imageReplacementPrefix("east"), "") + "/foo:bar"
			Expect(imageConfig.FindTargetImage(prefixSource+"/foo:bar", "local2", "east")).To(Equal(expectedTargetImageEast))
		})

		It("should find the target image with image replacement", func() {
			expectedTargetImageWest := ptr.Deref(imageReplacement("west"), "")
			Expect(imageConfig.FindTargetImage(image, "local", "west")).To(Equal(expectedTargetImageWest))

			expectedTargetImageEast := ptr.Deref(imageReplacement("east"), "")
			Expect(imageConfig.FindTargetImage(image, "local", "east")).To(Equal(expectedTargetImageEast))
		})

		It("should not find an image for an unknown provider, region", func() {
			Expect(imageConfig.FindTargetImage(image, "local3", "west")).To(BeEmpty())
			Expect(imageConfig.FindTargetImage(image, "local", "central")).To(BeEmpty())
		})
	})

	Describe("#HasOverwrite", func() {
		BeforeEach(func() {
			imageConfig = NewImageConfiguration(config)
		})

		It("should return true if an overwrite exists for the given image, provider, and region", func() {
			Expect(imageConfig.HasOverwrite("local", "west")).To(BeTrue())
			Expect(imageConfig.HasOverwrite("local", "east")).To(BeTrue())
		})

		It("should return false if no overwrite exists for the given image, provider, and region", func() {
			Expect(imageConfig.HasOverwrite("local3", "west")).To(BeFalse())
			Expect(imageConfig.HasOverwrite("local2", "central")).To(BeFalse())
		})
	})
})

func imageReplacementPrefix(region string) *string {
	return ptr.To("local2-" + region)
}

func imageReplacement(region string) *string {
	return ptr.To("local-" + region + "/image-replacement:latest")
}
