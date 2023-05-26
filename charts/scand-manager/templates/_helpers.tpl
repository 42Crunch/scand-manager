{{/*
Expand the name of the chart.
*/}}
{{- define "scand-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "scand-manager.fullname" -}}
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
{{- define "scand-manager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "scand-manager.labels" -}}
helm.sh/chart: {{ include "scand-manager.chart" . }}
{{ include "scand-manager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "scand-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "scand-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Extra Pod labels
*/}}
{{- define "scand-manager.extraPodLabels" -}}
{{- with .Values.extraPodLabels }}
{{- toYaml . }}
{{- end }}
{{- if .Values.firewallInjection.enabled }}
firewall-injection: enabled
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "scand-manager.serviceAccountName" -}}
{{- if .Values.rbac.enabled }}
{{- default (include "scand-manager.fullname" .) .Values.rbac.serviceAccount.nameOverride }}
{{- else }}
{{- default "default" .Values.rbac.serviceAccount.nameOverride }}
{{- end }}
{{- end }}

{{/*
Pod annotations
*/}}
{{- define "scand-manager.podAnnotations" -}}
{{- with .Values.podAnnotations }}
{{- toYaml . }}
{{- end }}
{{- if .Values.firewallInjection.enabled }}
xliic.com/protection-token: {{ .Values.firewallInjection.protectionToken }}
xliic.com/http-only: "enabled"
xliic.com/container-port: {{ .Values.firewallInjection.containerPort | quote }}
xliic.com/env-configmap: {{ .Values.firewallInjection.envConfigmap }}
xliic.com/target-url: {{ .Values.firewallInjection.targetUrl }}
xliic.com/server-name: {{ .Values.firewallInjection.serverName }}
{{- end }}
{{- end }}
