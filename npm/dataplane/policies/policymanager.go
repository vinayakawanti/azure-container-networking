package policies

import (
	"sync"

	"github.com/Azure/azure-container-networking/npm/api/v1"
)

type PolicyMap struct {
	cache map[string]*api.AclPolicy
	sync.Mutex
}

type PolicyManager struct {
	policyMap *PolicyMap
}

func NewPolicyManager() *PolicyManager {
	return &PolicyManager{
		policyMap: &PolicyMap{
			cache: make(map[string]*api.AclPolicy),
		},
	}
}
