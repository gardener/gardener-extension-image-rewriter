// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

//go:generate sh -c "bash $GARDENER_HACK_DIR/generate-crds.sh -p 20-crd- extensions.gardener.cloud resources.gardener.cloud"
//go:generate sh -c "find . -iname '20*' | grep -Ev '_clusters|_extensions|_managedresource' | xargs rm"
//go:generate sh -c "bash $GARDENER_HACK_DIR/generate-controller-registration.sh extension-image-rewriter ../charts/gardener-extension-image-rewriter $(cat ../VERSION) controller-registration.yaml Extension:image-rewriter"
//go:generate sh -c "kustomize build -o controller-registration.yaml"
//go:generate sh -c "extension-generator --name=extension-image-rewriter --provider-type=image-rewriter --component-category=extension --extension-oci-repository=europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/image-rewriter:$(cat ../VERSION) --destination=./extension/base/extension.yaml"
//go:generate sh -c "$TOOLS_BIN_DIR/kustomize build ./extension -o ./extension.yaml"

// Package example contains generated manifests for all CRDs and other examples.
// Useful for development purposes.
package example
