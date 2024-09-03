package tq

import (
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

func Build() (tqsdk.Space, error) {
	return tqsdk.Space{
		Key:    "key",
		Region: "nyc",
	}, nil
}
