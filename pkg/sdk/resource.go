package tqsdk

type Space struct {
	Name   string
	Region string

	Service Service
}

type Service struct {
	Key string
	// The path to a Dockerfile relative to the root of the repo. If set, overrides usage of buildpacks.
	DockerfilePath string
	BuildEnvs      map[string]string
	RuntimeEnvs    map[string]string
	BuildSecrets   map[string]string
	RuntimeSecrets map[string]string
	// The internal port on which this service's run command will listen.
	HttpPort int
	// An image to use as the component's source. Only one of `git`, `github`, `gitlab`, or `image` may be set.
	// Image AppSpecServiceImage
	// The amount of instances that this component should be scaled to.
	InstanceCount int
	// The instance size to use for this component. This determines the plan (basic or professional) and the available CPU and memory. The list of available instance sizes can be [found with the API](https://docs.digitalocean.com/reference/api/api-reference/#operation/list_instance_sizes) or using the [doctl CLI](https://docs.digitalocean.com/reference/doctl/) (`doctl apps tier instance-size list`). Default: `basic-xxs`
	InstanceSizeSlug string
	// The name of the component.
	Name     string
	SizeSlug SizeSlug
}
