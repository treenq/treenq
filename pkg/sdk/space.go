package tqsdk

type Space struct {
	Key    string
	Region string

	Service Service
}

type Service struct {
	Key string
	// The path to a Dockerfile relative to the root of the repo. If set, overrides usage of buildpacks.
	DockerfilePath string
	BuildEnvs      map[string]string
	RuntimeEnvs    map[string]string

	BuildSecrets   []string
	RuntimeSecrets []string
	// The internal port on which this service's run command will listen.
	HttpPort int
	// An image to use as the component's source. Only one of `git`, `github`, `gitlab`, or `image` may be set.
	// Image AppSpecServiceImage
	// Replicas defines the amount of instances that this component should be scaled to
	Replicas int
	// The name of the component.
	Name string
	// SizeSlug defines a compute resource requirement
	SizeSlug SizeSlug
	// ComputationResource is a verbose compute resource requirement, is mutual exclusive to SizeSlug
	ComputationResource ComputationResource
}
