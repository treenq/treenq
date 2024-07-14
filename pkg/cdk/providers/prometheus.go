package providers

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"text/template"
	"time"
)

type Job struct {
	JobName string
	Targets []string
}

type Config struct {
	Interval time.Duration
	Jobs     []Job
}

type ConfigBuilder struct {
	tpl *template.Template
}

//go:embed tpls/prometheus.tpl
var prometheusConfigTpl string

func NewConfigBuilder() (*ConfigBuilder, error) {
	goTpl, err := template.New("prometheusTpl").Parse(prometheusConfigTpl)
	if err != nil {
		return nil, err
	}

	return &ConfigBuilder{tpl: goTpl}, nil
}

func (b *ConfigBuilder) Build(conf Config) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)

	if err := b.tpl.Execute(buf, conf); err != nil {
		return nil, fmt.Errorf("failed to execute config template: %w", err)
	}

	return buf, nil
}

func (b *ConfigBuilder) BuildAsContent(conf Config) (string, error) {
	r, err := b.Build(conf)
	if err != nil {
		return "", err
	}
	return r.(*bytes.Buffer).String(), nil
}
