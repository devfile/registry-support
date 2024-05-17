#
# Copyright Red Hat
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

{{- define "devfileregistry.hostname" -}}
{{- .Values.hostnameOverride | default (printf "devfile-registry-%s" .Release.Namespace) -}}
{{- end -}}

{{- define "devfileregistry.ingressHostname" -}}
{{- $hostname := .Values.hostnameOverride | default (printf "devfile-registry-%s" .Release.Namespace) -}}
{{- .Values.global.ingress.domain | printf "%s.%s" $hostname -}}
{{- end -}}

{{- define "devfileregistry.routeHostname" -}}
{{- $hostname := .Values.hostnameOverride | default (printf "devfile-registry-%s" .Release.Namespace) -}}
{{- if eq .Values.global.route.domain "" -}} # This allows for Openshift to generate the route name + domain
{{- .Values.global.route.domain -}}
{{- else -}}
{{- .Values.global.route.domain | printf "%s.%s" $hostname -}}
{{- end -}}
{{- end -}}

{{- define "devfileregistry.ingressUrl" -}}
{{- $hostname := .Values.hostnameOverride | default (printf "devfile-registry-%s" .Release.Namespace) -}}
{{- if .Values.global.tlsEnabled -}}
{{- .Values.global.ingress.domain | printf "https://%s.%s" $hostname -}}
{{- else -}}
{{- .Values.global.ingress.domain | printf "http://%s.%s" $hostname -}}
{{- end -}}
{{- end -}}

{{- define "devfileregistry.routeUrl" -}}
{{- $hostname := .Values.hostnameOverride | default (printf "devfile-registry-%s" .Release.Namespace) -}}
{{- if .Values.global.tlsEnabled -}}
{{- .Values.global.route.domain | printf "https://%s.%s" $hostname -}}
{{- else -}}
{{- .Values.global.route.domain | printf "http://%s.%s" $hostname -}}
{{- end -}}
{{- end -}}

{{- define "devfileregistry.fqdnUrl" -}}
{{- if .Values.global.isOpenShift -}}
{{- template "devfileregistry.routeUrl" . -}}
{{- else -}}
{{- template "devfileregistry.ingressUrl" . -}}
{{- end -}}
{{- end -}}

{{- define "devfileregistry.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "devfileregistry.fullname" -}}
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
