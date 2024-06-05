package tqsdk

type Resource struct {
	App App
}

type App struct {
	Name string
	Git  AppGit

	Port         string
	BuildCommand string
	RunCommand   string
	Envs         map[string]string
}

type AppGit struct {
	Url    string
	Branch string
}

func (r Resource) AsConfig() Config {
	return Config{
		App: AppConfig{
			Host: "super host",
		},
	}
}

type Config struct {
	App AppConfig
}

type AppConfig struct {
	Host string
}
