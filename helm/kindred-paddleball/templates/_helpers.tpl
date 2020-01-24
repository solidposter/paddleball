{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "kindred-paddleball.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kindred-paddleball.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kindred-paddleball.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "kindred-paddleball.labels" -}}
app.kubernetes.io/name: {{ include "kindred-paddleball.name" . }}
helm.sh/chart: {{ include "kindred-paddleball.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.Version }}
app.kubernetes.io/version: {{ .Chart.Version | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Contaner args
*/}}
{{- define "kindred-paddleball.args" -}}
{{- if .Values.serverMode }}
["-k", {{ .Values.server.key | quote }}, "-s", {{ .Values.service.port | quote }}]
{{- else -}}
["-k", {{ .Values.client.key | quote }}, "{{ .Values.client.host }}:{{ .Values.client.port }}"]
{{- end }}
{{- end -}}