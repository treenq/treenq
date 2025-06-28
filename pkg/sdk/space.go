package tqsdk

import "errors"

var (
	ErrServiceNameRequired = errors.New("service.name required")
	ErrHttpPortRequired    = errors.New("service.httpPort required")
)

const (
	DefaultDockerfilePath = "Dockerfile"
	DefaultDockerContext  = "."
	DefaultReplicas       = 1
	DefaultCpuUnit        = 1000 // 1 cpu
	DefaultMemoryMibs     = 2048
	DefaultDiskGibs       = 20
)

type Space struct {
	Version string
	Service Service
}

type Service struct {
	// The name of the component,
	// unique in the space
	Name string `json:"name"`

	// The path to a Dockerfile relative to the root of the repo. If set, overrides usage of buildpacks.
	DockerfilePath string `json:"dockerfilePath"`
	// context for docker build
	DockerContext string `json:"dockerContext"`
	// runtime envs
	RuntimeEnvs map[string]string `json:"runtimeEnvs"`

	// The internal port on which this service's run command will listen.
	HttpPort int `json:"httpPort"`
	// Replicas defines the amount of instances that this component should be scaled to
	Replicas int `json:"replicas"`
	// ComputationResource is a verbose compute resource requirement, is mutual exclusive to SizeSlug
	ComputationResource ComputationResource `json:"computationResource"`
}

type ComputationResource struct {
	CpuUnits   int `json:"cpuUnits"`
	MemoryMibs int `json:"memoryMibs"`
	DiskGibs   int `json:"diskGibs"`
}

func (s *Space) Validate() error {
	if s.Service.Name == "" {
		return ErrServiceNameRequired
	}

	if s.Service.HttpPort == 0 {
		return ErrHttpPortRequired
	}

	if s.Service.DockerfilePath == "" {
		s.Service.DockerfilePath = DefaultDockerfilePath
	}
	if s.Service.DockerContext == "" {
		s.Service.DockerContext = DefaultDockerContext
	}
	if s.Service.Replicas <= 0 {
		s.Service.Replicas = DefaultReplicas
	}
	if s.Service.ComputationResource.CpuUnits <= 0 {
		s.Service.ComputationResource.CpuUnits = DefaultCpuUnit
	}
	if s.Service.ComputationResource.MemoryMibs <= 0 {
		s.Service.ComputationResource.MemoryMibs = DefaultMemoryMibs
	}
	if s.Service.ComputationResource.DiskGibs <= 0 {
		s.Service.ComputationResource.DiskGibs = DefaultDiskGibs
	}

	return nil
}
