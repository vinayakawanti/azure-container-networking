package dataplane

import (
	"runtime"

	"github.com/Azure/azure-container-networking/npm/dataplane/ipsets"
	"github.com/Azure/azure-container-networking/npm/dataplane/policies"
)

type OsType string

const (
	// Linux is the default OS type
	Linux   = OsType("linux")
	Windows = OsType("windows")
)

type DataPlane struct {
	policies.PolicyManager
	ipsets.IPSetManager
	OsType    OsType
	networkID string
	// key is PodKey
	endpointCache map[string]*Endpoint
}

func NewDataPlane() *DataPlane {
	return &DataPlane{
		OsType:        detectOsType(),
		PolicyManager: policies.NewPolicyManager(),
		IPSetManager:  ipsets.NewIPSetManager(string(detectOsType())),
	}
}

type DataPlane interface {
	InitializeDataplane() error
	ResetDataplane() error

	// ACLPolicy related functions
	// Add Policy takes in the custom NPMNetworkPolicy object
	// and adds it to the dataplane
	AddPolicies(policies *NPMNetworkPolicy) error
	// Delete Policy takes in name of the policy, looks up cache for policy obj
	// and deletes it from the dataplane
	RemovePolicies(policyName string) error
	// Update Policy takes in the custom NPMNetworkPolicy object
	// calculates the diff between the old and new policy and updates the dataplane
	UpdatePolicies(policies *NPMNetworkPolicy) error

	// IPSet related functions
	CreateIPSet(Set *IPSet) error
	DeleteSet(name string) error
	DeleteList(name string) error

	AddToSet(setNames []string, ip, podKey string) error
	RemoveFromSet(setNames []string, ip. podkey string) error
	AddToList(listName string, setNames []string) error
	RemoveFromList(listName string, setNames []string) error

	// UpdatePod helps in letting DP know about a new pod
	// this function will have two responsibilities,
	// 1. proactively get endpoint info of pod
	// 2. check if any of the existing policies applies to this pod
	//    and update ACLs on this pod's endpoint
	UpdatePod(pod *Pod) error

	// Called after all the ipsets operations are done
	// this call acts as a signal to the dataplane to update the kernel
	ApplyDataplane() error
}

// Detects the OS type
func detectOsType() OsType {
	os := runtime.GOOS
	switch os {
	case "linux":
		return Linux
	case "windows":
		return Windows
	default:
		panic("Unsupported OS type: " + os)
	}
}
