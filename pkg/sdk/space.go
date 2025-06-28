package tqsdk

import (
	"errors"
)

var (
	ErrServiceNameRequired = errors.New("service.name required")
	ErrHttpPortRequired    = errors.New("service.httpPort required")
	ErrReleaseOnRequired   = errors.New("service.releaseOn requires one of the fields")
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
	Service Service
}

type Service struct {
	// The name of the component,
	// unique in the space
	Name string `json:"name"`

	// ReleaseOn defines the release strategy on different merge events
	ReleaseOn ReleaseOn `json:"releaseOn"`

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

// ReleaseOn defines the release strategy on different merge events,
// only one of them must be defines
type ReleaseOn struct {
	// Branch requires a branch name
	Branch string
	// TagPrefix expects a tag prefix which will be listened,
	// "*" is allowed to specify any tag triggers a release
	TagPrefix string
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

	if s.Service.ReleaseOn.Branch == "" && s.Service.ReleaseOn.TagPrefix == "" {
		return ErrReleaseOnRequired
	}
	if s.Service.ReleaseOn.Branch != "" && s.Service.ReleaseOn.TagPrefix != "" {
		return ErrReleaseOnRequired
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
