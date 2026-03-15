{{/*
Expand the name of the chart.
*/}}
{{- define "mantl-platform.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "mantl-platform.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mantl-platform.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "mantl-platform.labels" -}}
helm.sh/chart: {{ include "mantl-platform.chart" . }}
{{ include "mantl-platform.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "mantl-platform.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mantl-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "mantl-platform.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "mantl-platform.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get the namespace
*/}}
{{- define "mantl-platform.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride }}
{{- end }}

{{/*
Return the appropriate apiVersion for RBAC APIs
*/}}
{{- define "mantl-platform.rbac.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "rbac.authorization.k8s.io/v1" }}
rbac.authorization.k8s.io/v1
{{- else }}
rbac.authorization.k8s.io/v1beta1
{{- end }}
{{- end }}

{{/*
Return the appropriate apiVersion for Ingress
*/}}
{{- define "mantl-platform.ingress.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "networking.k8s.io/v1" }}
networking.k8s.io/v1
{{- else if .Capabilities.APIVersions.Has "networking.k8s.io/v1beta1" }}
networking.k8s.io/v1beta1
{{- else }}
extensions/v1beta1
{{- end }}
{{- end }}

{{/*
Determine if Ingress is stable
*/}}
{{- define "mantl-platform.ingress.isStable" -}}
{{- eq (include "mantl-platform.ingress.apiVersion" .) "networking.k8s.io/v1" }}
{{- end }}

{{/*
Determine if Ingress supports IngressClassName
*/}}
{{- define "mantl-platform.ingress.supportsIngressClassName" -}}
{{- or (eq (include "mantl-platform.ingress.isStable" .) "true") (and (eq (include "mantl-platform.ingress.apiVersion" .) "networking.k8s.io/v1beta1")) }}
{{- end }}

{{/*
Determine if Ingress supports PathType
*/}}
{{- define "mantl-platform.ingress.supportsPathType" -}}
{{- or (eq (include "mantl-platform.ingress.isStable" .) "true") (and (eq (include "mantl-platform.ingress.apiVersion" .) "networking.k8s.io/v1beta1")) }}
{{- end }}

{{/*
Profile-based configuration merger
Returns merged configuration based on selected profile
*/}}
{{- define "mantl-platform.profileConfig" -}}
{{- $profile := .Values.profile | default "production" }}
{{- $profileValues := index .Values $profile | default dict }}
{{- $merged := merge (deepCopy $profileValues) (deepCopy .Values) }}
{{- toYaml $merged }}
{{- end }}

{{/*
Component enabled check with profile override
Usage: {{ include "mantl-platform.componentEnabled" (dict "root" . "component" "monitoring" "subcomponent" "prometheus") }}
*/}}
{{- define "mantl-platform.componentEnabled" -}}
{{- $root := .root }}
{{- $component := .component }}
{{- $subcomponent := .subcomponent | default "" }}
{{- $profile := $root.Values.profile | default "production" }}
{{- $profileValues := index $root.Values $profile | default dict }}
{{- $componentPath := index $root.Values $component | default dict }}
{{- $profileComponentPath := index $profileValues $component | default dict }}
{{- $enabled := false }}

{{- if $subcomponent }}
  {{- $subPath := index $componentPath $subcomponent | default dict }}
  {{- $profileSubPath := index $profileComponentPath $subcomponent | default dict }}
  {{- $enabled = or (index $profileSubPath "enabled") (index $subPath "enabled") }}
{{- else }}
  {{- $enabled = or (index $profileComponentPath "enabled") (index $componentPath "enabled") }}
{{- end }}

{{- $enabled }}
{{- end }}
