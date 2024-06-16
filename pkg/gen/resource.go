package gen

import (
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type App struct {
	ProjectID string
	Spec      AppSpec
}

type AppSpec struct {
	Name     string
	Region   string
	Services []AppSpecService
}

type AppSpecService struct {
	// Describes an alert policy for the component.
	// Alerts []AppSpecServiceAlert
	// An optional build command to run while building this component from source.
	BuildCommand string
	// The [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) policies of the app.
	// Deprecated: Service level CORS rules are deprecated in favor of ingresses
	Cors AppSpecServiceCors
	// The path to a Dockerfile relative to the root of the repo. If set, overrides usage of buildpacks.
	// DockerfilePath string
	// An environment slug describing the type of this app.
	EnvironmentSlug string
	// Describes an environment variable made available to an app competent.
	Envs []AppSpecServiceEnv
	// A Git repo to use as the component's source. The repository must be able to be cloned without authentication.  Only one of `git`, `github` or `gitlab`  may be set.
	// Git AppSpecServiceGit
	// A GitHub repo to use as the component's source. DigitalOcean App Platform must have [access to the repository](https://cloud.digitalocean.com/apps/github/install). Only one of `git`, `github`, `gitlab`, or `image` may be set.
	Github AppSpecServiceGithub
	// A Gitlab repo to use as the component's source. DigitalOcean App Platform must have [access to the repository](https://cloud.digitalocean.com/apps/gitlab/install). Only one of `git`, `github`, `gitlab`, or `image` may be set.
	// Gitlab AppSpecServiceGitlab
	// A health check to determine the availability of this component.
	HealthCheck AppSpecServiceHealthCheck
	// The internal port on which this service's run command will listen.
	HttpPort int
	// An image to use as the component's source. Only one of `git`, `github`, `gitlab`, or `image` may be set.
	// Image AppSpecServiceImage
	// The amount of instances that this component should be scaled to.
	InstanceCount int
	// The instance size to use for this component. This determines the plan (basic or professional) and the available CPU and memory. The list of available instance sizes can be [found with the API](https://docs.digitalocean.com/reference/api/api-reference/#operation/list_instance_sizes) or using the [doctl CLI](https://docs.digitalocean.com/reference/doctl/) (`doctl apps tier instance-size list`). Default: `basic-xxs`
	InstanceSizeSlug string
	// A list of ports on which this service will listen for internal traffic.
	InternalPorts []int
	// Describes a log forwarding destination.
	// LogDestinations []AppSpecServiceLogDestination
	// The name of the component.
	Name string
	// An HTTP paths that should be routed to this component.
	// Deprecated: Service level routes are deprecated in favor of ingresses
	Routes []AppSpecServiceRoute
	// An optional run command to override the component's default.
	RunCommand string
	// An optional path to the working directory to use for the build.
	SourceDir string
}

type AppSpecServiceAlert struct {
	// Determines whether or not the alert is disabled (default: `false`).
	Disabled bool
	// The operator to use. This is either of `GREATER_THAN` or `LESS_THAN`.
	Operator string
	// The type of the alert to configure. Component app alert policies can be: `CPU_UTILIZATION`, `MEM_UTILIZATION`, or `RESTART_COUNT`.
	Rule string
	// The threshold for the type of the warning.
	Value float64
	// The time before alerts should be triggered. This is may be one of: `FIVE_MINUTES`, `TEN_MINUTES`, `THIRTY_MINUTES`, `ONE_HOUR`.
	Window string
}

type AppSpecServiceCors struct {
	// Whether browsers should expose the response to the client-side JavaScript code when the request’s credentials mode is `include`. This configures the Access-Control-Allow-Credentials header.
	AllowCredentials bool
	// The set of allowed HTTP request headers. This configures the Access-Control-Allow-Headers header.
	AllowHeaders []string
	// The set of allowed HTTP methods. This configures the Access-Control-Allow-Methods header.
	AllowMethods []string
	// The set of allowed CORS origins. This configures the Access-Control-Allow-Origin header.
	AllowOrigins AppSpecServiceCorsAllowOrigins
	// The set of HTTP response headers that browsers are allowed to access. This configures the Access-Control-Expose-Headers header.
	ExposeHeaders []string
	// An optional duration specifying how long browsers can cache the results of a preflight request. This configures the Access-Control-Max-Age header. Example: `5h30m`.
	MaxAge string
}

type AppSpecServiceEnv struct {
	// The name of the environment variable.
	Key string
	// The visibility scope of the environment variable. One of `RUN_TIME`, `BUILD_TIME`, or `RUN_AND_BUILD_TIME` (default).
	Scope string
	// The type of the environment variable, `GENERAL` or `SECRET`.
	Type string
	// The value of the environment variable.
	Value string
}

type AppSpecServiceGit struct {
	// The name of the branch to use.
	Branch string
	// The clone URL of the repo.
	RepoCloneUrl string
}

type AppSpecServiceGithub struct {
	// The name of the branch to use.
	Branch string
	// Whether to automatically deploy new commits made to the repo.
	DeployOnPush bool
	// The name of the repo in the format `owner/repo`.
	Repo string
}

type AppSpecServiceGitlab struct {
	// The name of the branch to use.
	Branch string
	// Whether to automatically deploy new commits made to the repo.
	DeployOnPush bool
	// The name of the repo in the format `owner/repo`.
	Repo string
}

type AppSpecServiceHealthCheck struct {
	// The number of failed health checks before considered unhealthy.
	FailureThreshold int
	// The route path used for the HTTP health check ping.
	HttpPath string
	// The number of seconds to wait before beginning health checks.
	InitialDelaySeconds int
	// The number of seconds to wait between health checks.
	PeriodSeconds int
	// The health check will be performed on this port instead of component's HTTP port.
	Port int
	// The number of successful health checks before considered healthy.
	SuccessThreshold int
	// The number of seconds after which the check times out.
	TimeoutSeconds int
}

type AppSpecServiceImage struct {
	// Configures automatically deploying images pushed to DOCR.
	DeployOnPushes []AppSpecServiceImageDeployOnPush
	// The registry name. Must be left empty for the `DOCR` registry type. Required for the `DOCKER_HUB` registry type.
	Registry string
	// Access credentials for third-party registries
	RegistryCredentials string
	// The registry type. One of `DOCR` (DigitalOcean container registry) or `DOCKER_HUB`.
	RegistryType string
	// The repository name.
	Repository string
	// The repository tag. Defaults to `latest` if not provided.
	Tag string
}

type AppSpecServiceLogDestination struct {
	// Datadog configuration.
	Datadog AppSpecServiceLogDestinationDatadog
	// Logtail configuration.
	Logtail AppSpecServiceLogDestinationLogtail
	// Name of the log destination. Minimum length: 2. Maximum length: 42.
	Name string
	// Papertrail configuration.
	Papertrail AppSpecServiceLogDestinationPapertrail
}

type AppSpecServiceRoute struct {
	// Paths must start with `/` and must be unique within the app.
	Path string
	// An optional flag to preserve the path that is forwarded to the backend service.
	PreservePathPrefix bool
}

type AppSpecServiceCorsAllowOrigins struct {
	// Exact string match.
	Exact string
	// Prefix-based match.
	Prefix string
	// RE2 style regex-based match.
	Regex string
}

type AppSpecServiceImageDeployOnPush struct {
	// Whether to automatically deploy images pushed to DOCR.
	Enabled bool
}

type AppSpecServiceLogDestinationDatadog struct {
	// Datadog API key.
	ApiKey string
	// Datadog HTTP log intake endpoint.
	Endpoint string
}

type AppSpecServiceLogDestinationLogtail struct {
	// Logtail token.
	Token string
}

type AppSpecServiceLogDestinationPapertrail struct {
	// Papertrail syslog endpoint.
	Endpoint string
}

func MakeAppResourceData(app App) *schema.ResourceData {
	var d schema.ResourceData

	digitalocean.Provider()

	services := make([]map[string]any, len(app.Spec.Services))

	for i, service := range app.Spec.Services {
		envs := make([]map[string]any, len(service.Envs))
		for j, env := range service.Envs {
			envs[j] = map[string]any{
				"key":   env.Key,
				"value": env.Value,
				"scope": env.Scope,
				"type":  env.Type,
			}
		}

		routes := make([]map[string]any, len(service.Routes))
		for j, route := range service.Routes {
			routes[j] = map[string]any{
				"path":                 route.Path,
				"preserve_path_prefix": route.PreservePathPrefix,
			}
		}

		services[i] = map[string]any{
			"build_command": service.BuildCommand,
			"run_command":   service.RunCommand,
			"envionment_slug": map[string]any{
				"slug": service.EnvironmentSlug,
			},
			"http_port":          service.HttpPort,
			"instance_count":     service.InstanceCount,
			"instance_size_slug": service.InstanceSizeSlug,
			"internal_ports":     service.InternalPorts,
			"name":               service.Name,
			"source_dir":         service.SourceDir,
			"cors": map[string]any{
				"allow_credentials": service.Cors.AllowCredentials,
				"allow_headers":     service.Cors.AllowHeaders,
				"allow_methods":     service.Cors.AllowMethods,
				"expose_headers":    service.Cors.ExposeHeaders,
				"max_age":           service.Cors.MaxAge,
				"allow_origins": map[string]any{
					"exact":  service.Cors.AllowOrigins.Exact,
					"prefix": service.Cors.AllowOrigins.Prefix,
					"regex":  service.Cors.AllowOrigins.Regex,
				},
			},
			"envs": envs,
			"github": map[string]any{
				"branch":         service.Github.Branch,
				"deploy_on_push": service.Github.DeployOnPush,
				"repo":           service.Github.Repo,
			},
			"health_check": map[string]any{
				"failure_threshold":     service.HealthCheck.FailureThreshold,
				"http_path":             service.HealthCheck.HttpPath,
				"initial_delay_seconds": service.HealthCheck.InitialDelaySeconds,
				"period_seconds":        service.HealthCheck.PeriodSeconds,
				"port":                  service.HealthCheck.Port,
				"success_threshold":     service.HealthCheck.SuccessThreshold,
				"timeout_seconds":       service.HealthCheck.TimeoutSeconds,
			},
			"routes": routes,
		}
	}

	value := []map[string]any{
		{
			"project_id": app.ProjectID,
			"spec": map[string]any{
				"name":     app.Spec.Name,
				"region":   app.Spec.Region,
				"services": services,
			},
		},
	}

	d.Set("spec", value)

	return &d
}
