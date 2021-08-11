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

func (pMgr *PolicyManager) Get(name string) (*api.AclPolicy, error) {
	pMgr.policyMap.Lock()
	defer pMgr.policyMap.Unlock()

	if policy, ok := pMgr.policyMap.cache[name]; ok {
		return policy, nil
	}

	return nil, nil
}

func (pMgr *PolicyManager) Add(policy *api.AclPolicy) error {
	pMgr.policyMap.Lock()
	defer pMgr.policyMap.Unlock()

	// actually apply on the dataplane
	// here we assume most of the logic will be done
	// by the os specific dataplane

	pMgr.policyMap.cache[policy.PolicyName] = policy
	return nil
}

func (pMgr *PolicyManager) Remove(name string) error {
	pMgr.policyMap.Lock()
	defer pMgr.policyMap.Unlock()

	if _, ok := pMgr.policyMap.cache[name]; !ok {
		return nil
	}

	// actually apply on the dataplane
	// here we assume most of the logic will be done
	// by the os specific dataplane

	delete(pMgr.policyMap.cache, name)

	return nil
}
