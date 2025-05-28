// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"github.com/gardener/gardener-extension-image-rewriter/pkg/image"
)

type fakeImageConfiguration struct {
	sourceToTarget map[string]string
	hasOverwrite   bool
}

var _ image.Configuration = (*fakeImageConfiguration)(nil)

func (f *fakeImageConfiguration) FindTargetImage(source, _, _ string) string {
	return f.sourceToTarget[source]
}

func (f *fakeImageConfiguration) HasOverwrite(_, _ string) bool {
	return f.hasOverwrite
}

// NewFakeImageConfiguration creates a new fake image configuration with the given source-to-target mapping and overwrite flag.
func NewFakeImageConfiguration(sourceToTarget map[string]string, hasOverwrite bool) image.Configuration {
	return &fakeImageConfiguration{
		sourceToTarget: sourceToTarget,
		hasOverwrite:   hasOverwrite,
	}
}
