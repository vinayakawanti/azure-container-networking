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
	OsType    OsType
	policyMgr *policies.PolicyManager
	setMgr    *ipsets.IPSetManager
}

func NewDataPlane() *DataPlane {
	return &DataPlane{
		OsType:    detectOsType(),
		policyMgr: policies.NewPolicyManager(),
		setMgr:    ipsets.NewIPSetManager(),
	}
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
