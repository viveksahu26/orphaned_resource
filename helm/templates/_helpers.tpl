{{/*
Expand the name of the chart.
*/}}
{{- define "obmondo-k8s-agent.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "obmondo-k8s-agent.fullname" -}}
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
{{- define "obmondo-k8s-agent.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "obmondo-k8s-agent.labels" -}}
helm.sh/chart: {{ include "obmondo-k8s-agent.chart" . }}
{{ include "obmondo-k8s-agent.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "obmondo-k8s-agent.selectorLabels" -}}
app.kubernetes.io/name: {{ include "obmondo-k8s-agent.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "obmondo-k8s-agent.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "obmondo-k8s-agent.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{/* EXPLAINATION:

This is a set of Helm chart templates for creating Kubernetes resources. 
Here's a breakdown of each template and what it does:

1. obmondo-k8s-agent.name: This template defines a function 
that expands the name of the chart. 
It uses the .Chart.Name and .Values.nameOverride 
variables to create a name for the chart. 
If .Values.nameOverride is not set, it will use the chart name. 
The result is truncated to 63 characters, with any trailing dashes removed.

2. obmondo-k8s-agent.fullname: This template defines a function 
that creates a fully qualified app name. 
It uses .Values.fullnameOverride, .Chart.Name, 
and .Release.Name to generate the name. 
If .Values.fullnameOverride is set, it will use that value. 
Otherwise, it concatenates the release name and chart name 
with a hyphen. If the resulting name is longer than 63 
characters, it will be truncated to 63 characters and any 
trailing dashes will be removed.

3. obmondo-k8s-agent.chart: This template defines a function 
that creates the chart name and version for use in the chart 
label. It concatenates .Chart.Name and .Chart.Version, 
replaces any + characters with _, and truncates the result to 
63 characters with any trailing dashes removed.

4.obmondo-k8s-agent.labels: This template defines a function 
that creates a set of common labels for use in Kubernetes 
resources. It uses the helm.sh/chart label, 
the obmondo-k8s-agent.selectorLabels function to create 
additional selector labels, and, if .Chart.AppVersion is set, 
an app.kubernetes.io/version label. It also sets an app.kubernetes.io/managed-by 
label to the value of .Release.Service.

5. obmondo-k8s-agent.selectorLabels: This template defines a 
function that creates a set of selector labels for use in 
Kubernetes resources. It uses the obmondo-k8s-agent.name 
function to create an app.kubernetes.io/name label and 
.Release.Name to create an app.kubernetes.io/instance label.

6.obmondo-k8s-agent.serviceAccountName: This template defines a 
function that creates the name of the service account to use in 
the chart. It uses the .Values.serviceAccount.create variable to
 determine whether to create a new service account or use an 
 existing one. If .Values.serviceAccount.create is true, 
 it uses the obmondo-k8s-agent.fullname function to create the 
 service account name. Otherwise, it uses the 
 .Values.serviceAccount.name variable to specify the name of 
 the existing service account to use.
*/}}
