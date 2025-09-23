// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package containerd

import (
	"bytes"
	_ "embed"
	"regexp"
	"text/template"
)

var (
	//go:embed templates/hosts.toml.tpl
	tplContentHosts string
	tplHosts        *template.Template

	hostWithPathPattern = regexp.MustCompile(`https?://[a-zA-Z0-9\.\-]+(/[^\s]*)+`)
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
}

// HostsTOML returns hosts.toml configuration.
func (r *RegistryMirror) HostsTOML() (string, error) {
	values := map[string]any{
		"server": r.UpstreamServer,
		"host":   r.MirrorHost,
	}

	// If the host URL contains a path, override_path needs to be set to true, see https://github.com/containerd/containerd/blob/main/docs/hosts.md#override_path-field.
	if hostWithPathPattern.MatchString(r.MirrorHost) {
		values["overridePath"] = true
	}

	hostsTOML := bytes.NewBuffer(nil)

	err := tplHosts.Execute(hostsTOML, values)
	if err != nil {
		return "", err
	}

	return hostsTOML.String(), nil
}
