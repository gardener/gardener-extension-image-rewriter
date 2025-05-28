# Gardener Extension for Image Rewrites

This project provides a Gardener extension that replaces pod image references for **system components** of shoot clusters.

## Components

- Mutating webhook for shoots to replace image references of `Pod`s running in the `kube-system` namespace.
- Mutating webhook for seeds to replace image references in `OperatingSystemConfig` resources.

## Configuration

A separate configuration file is used to define the image rewrites. Please see [here](./example/00-componentconfig.yaml) for an example.
