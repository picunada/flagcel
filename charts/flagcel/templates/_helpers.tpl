{{/* Expand the name of the chart. */}}
{{- define "flagcel.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Fully qualified app name. */}}
{{- define "flagcel.fullname" -}}
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

{{- define "flagcel.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "flagcel.labels" -}}
helm.sh/chart: {{ include "flagcel.chart" . }}
{{ include "flagcel.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "flagcel.selectorLabels" -}}
app.kubernetes.io/name: {{ include "flagcel.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "flagcel.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "flagcel.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/* Image reference, defaulting the tag to the chart appVersion. */}}
{{- define "flagcel.image" -}}
{{- printf "%s:%s" .Values.image.repository (default .Chart.AppVersion .Values.image.tag) }}
{{- end }}

{{/* Resolve the DATABASE_URL written into the chart-managed Secret. */}}
{{- define "flagcel.databaseUrl" -}}
{{- if .Values.database.url -}}
{{- .Values.database.url -}}
{{- else if .Values.postgresql.enabled -}}
{{- printf "postgres://%s:%s@%s-postgresql:5432/%s?sslmode=disable" .Values.postgresql.auth.username .Values.postgresql.auth.password .Release.Name .Values.postgresql.auth.database -}}
{{- else -}}
{{- fail "flagcel: set database.url, database.existingSecret, or postgresql.enabled" -}}
{{- end -}}
{{- end -}}

{{/* Session secret: explicit value, else stable existing value, else random. */}}
{{- define "flagcel.sessionSecret" -}}
{{- if .Values.auth.sessionSecret -}}
{{- .Values.auth.sessionSecret -}}
{{- else -}}
{{- $existing := lookup "v1" "Secret" .Release.Namespace (include "flagcel.fullname" .) -}}
{{- if and $existing $existing.data (index $existing.data "AUTH_SESSION_SECRET") -}}
{{- index $existing.data "AUTH_SESSION_SECRET" | b64dec -}}
{{- else -}}
{{- randAlphaNum 48 -}}
{{- end -}}
{{- end -}}
{{- end -}}
