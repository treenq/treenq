package tqsdk

type Space struct {
	Version string
	Service Service
}

type Service struct {
	// The name of the component,
	// unique in the space
	Name string

	// The path to a Dockerfile relative to the root of the repo. If set, overrides usage of buildpacks.
	DockerfilePath string
	// context for docker build
	DockerContext string
	// runtime envs
	RuntimeEnvs map[string]string

	// The internal port on which this service's run command will listen.
	HttpPort int
	// Replicas defines the amount of instances that this component should be scaled to
	Replicas int
	// ComputationResource is a verbose compute resource requirement, is mutual exclusive to SizeSlug
	ComputationResource ComputationResource
}
