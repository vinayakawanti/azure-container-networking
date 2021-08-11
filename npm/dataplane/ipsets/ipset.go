package ipsets

import (
	"github.com/Azure/azure-container-networking/npm/api/v1"
	"github.com/Azure/azure-container-networking/npm/util"
)

type SetKind string

const (
	ListSet SetKind = "list"
	HashSet SetKind = "set"
)

func NewIPSet(name string, setType api.SetType) *api.IPSet {
	return &api.IPSet{
		Name:       name,
		HashedName: util.GetHashedName(name),
		IpPodKey:   make(map[string]string),
		Type:       setType,
		ReferCount: int32(0),
		Size:       int32(0), // Do we need this ? may be max limit
		IPSet:      make(map[string]*api.IPSet),
	}
}

func IncReferCount(set *api.IPSet) {
	set.ReferCount++
}

func DecReferCount(set *api.IPSet) {
	set.ReferCount--
}

func getSetKind(set *api.IPSet) SetKind {
	switch set.Type {
	case api.SetType_CIDRBlocks:
		return HashSet
	case api.SetType_NameSpace:
		return HashSet
	case api.SetType_NamedPorts:
		return HashSet
	case api.SetType_KeyLabelOfPod:
		return HashSet
	case api.SetType_KeyValueLabelOfPod:
		return HashSet
	case api.SetType_KeyLabelOfNameSpace:
		return ListSet
	case api.SetType_KeyValueLabelOfNameSpace:
		return ListSet
	case api.SetType_NestedLabelOfPod:
		return ListSet
	default:
		return "unknown"
	}
}
