// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	. "github.com/gardener/gardener-extension-image-rewriter/pkg/utils/containerd"
)

var _ = Describe("Containerd", func() {
	Describe("Configuration", func() {
		var (
			containerdConfig Configuration
			config           *v1alpha1.Configuration
		)

		BeforeEach(func() {
			config = &v1alpha1.Configuration{
				Containerd: []v1alpha1.ContainerdConfiguration{
					{
						Upstream: "upstream1",
						Server:   "https://server1",
						Hosts: []v1alpha1.ContainerdHostConfig{
							{URL: "https://mirror1-west", Provider: "local", Regions: []string{"west"}},
							{URL: "https://mirror1-central", Provider: "local", Regions: []string{"central", "south", "north"}},
							{URL: "https://mirror1-east", Provider: "local", Regions: []string{"east"}},
						},
					},
					{
						Upstream: "upstream2",
						Server:   "https://server2",
						Hosts: []v1alpha1.ContainerdHostConfig{
							{URL: "https://mirror2-west", Provider: "local", Regions: []string{"west"}},
							{URL: "https://mirror2-central", Provider: "local", Regions: []string{"central", "south", "north"}},
							{URL: "https://mirror2-east", Provider: "local", Regions: []string{"east"}},
						},
					},
					{
						Upstream: "upstream3",
						Server:   "https://server3",
						Hosts: []v1alpha1.ContainerdHostConfig{
							{URL: "https://mirror3/west", Provider: "local2", Regions: []string{"west"}},
							{URL: "https://mirror3/central", Provider: "local2", Regions: []string{"central", "south", "north"}},
							{URL: "https://mirror3/east", Provider: "local2", Regions: []string{"east"}},
						},
					},
				},
			}

			containerdConfig = NewConfiguration(config)
		})

		Describe("#GetUpstreamConfig", func() {
			test := func(provider, region string, upstreamConfigs []UpStreamConfiguration) {
				GinkgoHelper()

				result := containerdConfig.GetUpstreamConfig(provider, region)

				Expect(result).To(HaveLen(len(upstreamConfigs)))

				for i, expected := range upstreamConfigs {
					Expect(result[i].Upstream).To(Equal(expected.Upstream), "upstream should match")
					Expect(result[i].Server).To(Equal(expected.Server), "server should match")
					Expect(result[i].HostURL).To(Equal(expected.HostURL), "host URL should match")
					Expect(result[i].OverridePath).To(Equal(expected.OverridePath), "override path should match")
				}
			}

			It("should return the correct upstream configuration for provision", func() {
				test("local", "west", []UpStreamConfiguration{
					{Upstream: "upstream1", Server: "https://server1", HostURL: "https://mirror1-west"},
					{Upstream: "upstream2", Server: "https://server2", HostURL: "https://mirror2-west"},
				})

				test("local", "central", []UpStreamConfiguration{
					{Upstream: "upstream1", Server: "https://server1", HostURL: "https://mirror1-central"},
					{Upstream: "upstream2", Server: "https://server2", HostURL: "https://mirror2-central"},
				})

				test("local", "south", []UpStreamConfiguration{
					{Upstream: "upstream1", Server: "https://server1", HostURL: "https://mirror1-central"},
					{Upstream: "upstream2", Server: "https://server2", HostURL: "https://mirror2-central"},
				})

				test("local", "north", []UpStreamConfiguration{
					{Upstream: "upstream1", Server: "https://server1", HostURL: "https://mirror1-central"},
					{Upstream: "upstream2", Server: "https://server2", HostURL: "https://mirror2-central"},
				})

				test("local", "east", []UpStreamConfiguration{
					{Upstream: "upstream1", Server: "https://server1", HostURL: "https://mirror1-east"},
					{Upstream: "upstream2", Server: "https://server2", HostURL: "https://mirror2-east"},
				})

				test("local2", "east", []UpStreamConfiguration{
					{Upstream: "upstream3", Server: "https://server3", HostURL: "https://mirror3/east", OverridePath: ptr.To(true)},
				})
			})

			It("should not find any configuration", func() {
				containerdConfig = NewConfiguration(&v1alpha1.Configuration{})

				test("local", "west", []UpStreamConfiguration{})

				test("local2", "central", []UpStreamConfiguration{})
			})

			It("should not find any configuration for unknown providers and regions", func() {
				test("local", "east-west", []UpStreamConfiguration{})

				test("local2", "south-central", []UpStreamConfiguration{})

				test("local3", "west", []UpStreamConfiguration{})
			})
		})
	})
})
