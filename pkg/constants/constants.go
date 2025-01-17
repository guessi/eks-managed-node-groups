package constants

const (
	AppName  string = "eks-managed-node-groups"
	AppUsage string = "managed Amazon EKS node groups made easy"
)

type NodeGroupType int

const (
	Managed NodeGroupType = iota
	SelfManaged
)

var NodeGroupTypes = map[NodeGroupType]string{
	Managed:     "Managed Node Group",
	SelfManaged: "Self-managed Node Group",
}

var (
	GitVersion string
	GoVersion  string
	BuildTime  string
)
