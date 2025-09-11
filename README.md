# Gardener Extension for Image Rewrites

[![REUSE status](https://api.reuse.software/badge/github.com/gardener/gardener-extension-image-rewriter)](https://api.reuse.software/info/github.com/gardener/gardener-extension-image-rewriter)

This project provides a Gardener extension that replaces pod image references for **system components** of shoot clusters.
Additionally, it can handle mirror configuration for containerd.

## Components

- Mutating webhook for shoots to replace image references of `Pod`s running in the `kube-system` namespace.
- Mutating webhook for seeds to replace image references in `OperatingSystemConfig` resources.
- Mutating webhook for seeds to add containerd configuration to `OperatingSystemConfig` resources.

## Configuration

A separate configuration file is used to define the image rewrites. Please see [here](./example/00-componentconfig.yaml) for an example.
