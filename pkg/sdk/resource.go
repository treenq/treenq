package tqsdk

type Resource struct {
	SpaceKey string
	Key      string
	Kind     ResourceKind
	Payload  any
}

type ResourceKind string

const (
	ResourceKindService ResourceKind = "service"
)
