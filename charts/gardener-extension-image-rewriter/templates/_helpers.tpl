{{- define "name" -}}
gardener-extension-image-rewriter
{{- end -}}

{{- define "config" -}}
apiVersion: config.image-rewriter.extensions.gardener.cloud/v1alpha1
kind: Configuration
{{- if .Values.overwrites }}
overwrites:
{{ toYaml .Values.overwrites | indent 2 }}
{{- end }}
{{- if .Values.containerd }}
containerd:
{{ toYaml .Values.containerd | indent 2 }}
{{- end }}
{{- end -}}

{{- define "configmap" -}}
{{- end }}

{{- define "leaderelectionid" -}}
extension-image-rewriter-leader-election
{{- end -}}

{{- define "image" -}}
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

{{- define "disabledcontrollers" }}
{{- if not .Values.overwrites -}}
image-rewriter-cluster-controller
{{- end }}
{{- end }}

{{- define "disabledwebhooks" }}
{{- $disabledWebhooks := list }}
{{- if not .Values.overwrites }}
{{- $disabledWebhooks = append $disabledWebhooks "pod-image-rewriter" }}
{{- $disabledWebhooks = append $disabledWebhooks "osc-image-rewriter" }}
{{- end }}
{{- if not .Values.containerd }}
{{- $disabledWebhooks = append $disabledWebhooks "osc-containerd" }}
{{- end }}
{{- join "," $disabledWebhooks -}}
{{- end -}}
