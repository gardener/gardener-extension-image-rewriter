// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd

import (
	"bytes"
	_ "embed"
	"text/template"
)

var (
	//go:embed templates/hosts.toml.tpl
	tplContentHosts string
	tplHosts        *template.Template
)

func init() {
	tplHosts = template.Must(template.
		New("hosts.toml").
		Parse(tplContentHosts))
}

// RegistryMirror represents a registry mirror for containerd.
type RegistryMirror struct {
	UpstreamServer string
	MirrorHost     string
	OverridePath   *bool
}

// HostsTOML returns hosts.toml configuration.
func (r *RegistryMirror) HostsTOML() (string, error) {
	values := map[string]any{
		"server": r.UpstreamServer,
		"host":   r.MirrorHost,
	}

	if r.OverridePath != nil {
		values["overridePath"] = *r.OverridePath
	}

	hostsTOML := bytes.NewBuffer(nil)

	err := tplHosts.Execute(hostsTOML, values)
	if err != nil {
		return "", err
	}

	return hostsTOML.String(), nil
}
