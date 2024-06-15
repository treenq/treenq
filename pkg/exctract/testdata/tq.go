package tq

import (
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

func Build() (tqsdk.Resource, error) {
	return tqsdk.Resource{
		App: tqsdk.App{
			Name:         "name",
			Port:         ":8000",
			BuildCommand: "go build some/thing.go",
			RunCommand:   "./thing",
		},
	}, nil
}
