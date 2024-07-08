package tq

import (
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

func Build() (tqsdk.Space, error) {
	return tqsdk.Space{
		Name:   "treenq-poc",
		Region: "nyc",
		Service: tqsdk.Service{
			SizeSlug:       tqsdk.SizeSlugs1vcpu512mb10gb,
			DockerfilePath: "Dockerfile",
			Name:           "treenq-poc-service",
			HttpPort:       8080,
			InstanceCount:  1,
		},
	}, nil
}
