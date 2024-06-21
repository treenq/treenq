package tq

import (
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

func Build() (tqsdk.App, error) {
	return tqsdk.App{
		Name:   "name",
		Region: "nyc",
	}, nil
}
