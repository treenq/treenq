package tq

import tqsdk "github.com/treenq/treenq/pkg/sdk"

func Build() (tqsdk.Resource, error) {
	git_config := tqsdk.AppGit{
		Url:    "some-url",
		Branch: "main",
	}
	envs := make(map[string]string)
	envs["is_prod"] = "false"
	app := tqsdk.App{
		Name:         "some-name",
		Port:         "8080",
		BuildCommand: "npm run",
		RunCommand:   "npm run",
		Git:          git_config,
		Envs:         envs,
	}
	return tqsdk.Resource{
		App: app,
	}, nil
}
