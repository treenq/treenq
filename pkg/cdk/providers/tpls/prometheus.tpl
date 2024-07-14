global:
  scrape_interval: {{ .Interval }}
  evaluation_interval: {{ .Interval }}

alerting:

rule_files:

scrape_configs:
{{- range .Jobs }}
  - job_name: {{ .JobName }}
    static_configs:
      - targets: 
        {{- range .Targets }}
        - "{{ . }}" {{- end }}
{{- end }}
