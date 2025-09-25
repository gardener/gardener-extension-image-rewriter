// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	. "github.com/gardener/gardener-extension-image-rewriter/pkg/utils/containerd"
)

var _ = Describe("RegistryMirror", func() {
	Describe("#HostsTOML", func() {
		It("generates correct configuration for valid fields", func() {
			mirror := RegistryMirror{
				UpstreamServer: "https://upstream.example.com",
				MirrorHost:     "https://mirror.example.com",
			}
			expected := `server = "https://upstream.example.com"

[host."https://mirror.example.com"]
  capabilities = ["pull", "resolve"]
`
			Expect(mirror.HostsTOML()).To(Equal(expected))
		})

		It("handles empty fields gracefully", func() {
			mirror := RegistryMirror{}
			expected := `server = ""

[host.""]
  capabilities = ["pull", "resolve"]
`
			Expect(mirror.HostsTOML()).To(Equal(expected))
		})

		It("handles hosts with path", func() {
			mirror := RegistryMirror{
				UpstreamServer: "https://upstream.example.com",
				MirrorHost:     "https://mirror.example.com/v2/some/path",
				OverridePath:   ptr.To(true),
			}
			expected := `server = "https://upstream.example.com"

[host."https://mirror.example.com/v2/some/path"]
  capabilities = ["pull", "resolve"]
  override_path = true
`
			Expect(mirror.HostsTOML()).To(Equal(expected))
		})
	})
})
