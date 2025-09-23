server = "{{ .server }}"

[host."{{ .host }}"]
  capabilities = ["pull", "resolve"]
{{- if .overridePath }}
  override_path = {{ .overridePath }}
{{- end }}
