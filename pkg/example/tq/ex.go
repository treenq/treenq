package tq

import tqsdk "github.com/treenq/treenq/pkg/sdk"

func Build() (tqsdk.Resource, error) {
	return tqsdk.Resource{
		Size: 1,
	}, nil
}
