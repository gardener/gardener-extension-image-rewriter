// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gardener/gardener-extension-image-rewriter/pkg/utils/containerd"
)

var _ = Describe("RegistryMirror", func() {
	Describe("#HostsTOML", func() {
		It("generates correct configuration for valid fields", func() {
			mirror := RegistryMirror{
				UpstreamServer: "https://upstream.example.com",
				MirrorHost:     "mirror.example.com",
			}
			expected := `server = "https://upstream.example.com"

[host."mirror.example.com"]
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
	})
})
