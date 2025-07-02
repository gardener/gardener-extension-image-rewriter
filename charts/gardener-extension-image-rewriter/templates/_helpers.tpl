{{- define "name" -}}
gardener-extension-image-rewriter
{{- end -}}

{{- define "config" -}}
apiVersion: config.image-rewriter.extensions.gardener.cloud/v1alpha1
kind: Configuration
overwrites:
{{ toYaml .Values.overwrites | indent 2 }}
containerd:
{{ toYaml .Values.containerd | indent 2 }}
{{- end -}}

{{- define "configmap" -}}
{{- end }}

{{- define "leaderelectionid" -}}
extension-image-rewriter-leader-election
{{- end -}}

{{-  define "image" -}}
  {{- if .Values.image.ref -}}
  {{ .Values.image.ref }}
  {{- else -}}
  {{- if hasPrefix "sha256:" .Values.image.tag }}
  {{- printf "%s@%s" .Values.image.repository .Values.image.tag }}
  {{- else }}
  {{- printf "%s:%s" .Values.image.repository .Values.image.tag }}
  {{- end }}
  {{- end -}}
{{- end }}
