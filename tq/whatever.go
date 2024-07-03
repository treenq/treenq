package tq

import (
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

func Build() (tqsdk.App, error) {
	return tqsdk.App{
		Name:     "treenq-poc",
		Region:   "nyc",
		SizeSlug: tqsdk.SizeSlugs1vcpu512mb10gb,
		Service: tqsdk.Service{
			DockerfilePath: "Dockerfile",
			Name:           "treenq-poc-service",
			HttpPort:       8080,
			InstanceCount:  1,
		},
	}, nil
}
