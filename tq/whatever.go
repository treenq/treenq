package tq

import (
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

func Build() (tqsdk.App, error) {
	return tqsdk.App{
		Name:     "treenq-poc",
		Region:   "FRA",
		SizeSlug: "basic-xxs",
		Service: tqsdk.Service{
			DockerfilePath: "Dockerfile",
			Name:           "treenq-poc-service",
			HttpPort:       8000,
			InstanceCount:  1,
		},
	}, nil
}
