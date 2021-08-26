// +build windows

package policies

import (
	"github.com/Azure/azure-container-networking/npm/api/v1"

	"github.com/Microsoft/hcsshim"
	"github.com/Microsoft/hcsshim/hcn"
)

/*
// AclPolicySetting creates firewall rules on an endpoint
type AclPolicySetting struct {
	Protocols       string        `json:",omitempty"` // EX: 6 (TCP), 17 (UDP), 1 (ICMPv4), 58 (ICMPv6), 2 (IGMP)
	Action          ActionType    `json:","`
	Direction       DirectionType `json:","`
	LocalAddresses  string        `json:",omitempty"`
	RemoteAddresses string        `json:",omitempty"`
	LocalPorts      string        `json:",omitempty"`
	RemotePorts     string        `json:",omitempty"`
	RuleType        RuleType      `json:",omitempty"`
	Priority        uint16        `json:",omitempty"`
}
*/

func convertPolicyToACL(policy *api.AclPolicy) (acl *hcn.AclPolicySetting) {
	acl = &hcn.AclPolicySetting{}

	acl.Action = hcn.ActionType(policy.Verdict)

	return
}
