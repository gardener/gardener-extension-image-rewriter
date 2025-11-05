// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	. "github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/validation"
)

var _ = Describe("Validation", func() {
	var config *v1alpha1.Configuration

	BeforeEach(func() {
		config = &v1alpha1.Configuration{
			Overwrites: []v1alpha1.ImageOverwrite{{}},
		}
	})

	Describe("#ValidateConfiguration", func() {
		It("should validate source and target are set", func() {
			Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("overwrites[0].source"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("overwrites[0].targets"),
			}))))
		})

		It("should validate only image or prefix is used for source", func() {
			config.Overwrites[0].Source = v1alpha1.Image{
				Image:  ptr.To("foo/bar:latest"),
				Prefix: ptr.To("foo"),
			}

			Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeForbidden),
				"Field": Equal("overwrites[0].source"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("overwrites[0].targets"),
			}))))
		})

		It("should validate target has required fields", func() {
			config.Overwrites[0].Source.Image = ptr.To("foo/bar:latest")
			config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{}}

			Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("overwrites[0].targets[0].image"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("overwrites[0].targets[0].provider"),
			}))))
		})

		It("should validate only prefix is configured", func() {
			config.Overwrites[0].Source.Prefix = ptr.To("foo")
			config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{
				Image: v1alpha1.Image{
					Image: ptr.To("foo/bar:latest"),
				},
				Provider: "local",
				Regions:  []string{"local"},
			}}

			Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("overwrites[0].targets[0].prefix"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeForbidden),
				"Field": Equal("overwrites[0].targets[0].image"),
			}))))
		})

		It("should validate only image is configured", func() {
			config.Overwrites[0].Source.Image = ptr.To("foo/bar:latest")
			config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{
				Image: v1alpha1.Image{
					Prefix: ptr.To("foo"),
				},
				Provider: "local",
				Regions:  []string{"local"},
			}}

			Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("overwrites[0].targets[0].image"),
			})), PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeForbidden),
				"Field": Equal("overwrites[0].targets[0].prefix"),
			}))))
		})

		It("should validate region is not empty", func() {
			config.Overwrites[0].Source.Image = ptr.To("foo/bar:latest")
			config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{
				Image: v1alpha1.Image{
					Image: ptr.To("foo/bar:latest"),
				},
				Provider: "local",
				Regions:  []string{""},
			}}

			Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("overwrites[0].targets[0].regions[0]"),
			}))))
		})
	})
})
