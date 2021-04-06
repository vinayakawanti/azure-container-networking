package network

import (
	"fmt"
	"net"
	"testing"

	"github.com/Azure/azure-container-networking/cni"
	"github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/network"
	"github.com/Azure/azure-container-networking/nns"
	"github.com/Azure/azure-container-networking/telemetry"
	cniSkel "github.com/containernetworking/cni/pkg/skel"
)

// the Add/Delete methods in Plugin require refactoring to have UT's written for them,
// but the mocks in this test are a start
func TestPlugin(t *testing.T) {
	config := &common.PluginConfig{}
	pluginName := "testplugin"

	//mockNetworkManager := acnnetwork.NewMockNetworkmanager()

	grpcClient := &nns.MockGrpcClient{}
	plugin, _ := NewPlugin(pluginName, config, grpcClient)
	plugin.report = &telemetry.CNIReport{}
	var err error
	plugin.nm = network.NewMockNetworkmanager()
	/*
		if err != nil {
			t.Fatal(err)
		}
	*/

	nwCfg := cni.NetworkConfig{
		Name:              "test-nwcfg",
		Type:              "azure-vnet",
		Mode:              "bridge",
		IPsToRouteViaHost: []string{"169.254.20.10"},
		Ipam: struct {
			Type          string `json:"type"`
			Environment   string `json:"environment,omitempty"`
			AddrSpace     string `json:"addressSpace,omitempty"`
			Subnet        string `json:"subnet,omitempty"`
			Address       string `json:"ipAddress,omitempty"`
			QueryInterval string `json:"queryInterval,omitempty"`
		}{
			Type: "azure-cns",
		},
	}

	args := &cniSkel.CmdArgs{
		ContainerID: "test-container",
		Netns:       "test-container",
	}
	args.StdinData = nwCfg.Serialize()
	podEnv := cni.K8SPodEnvArgs{
		K8S_POD_NAME:      "test-pod",
		K8S_POD_NAMESPACE: "test-pod-namespace",
	}
	args.Args = fmt.Sprintf("K8S_POD_NAME=%v;K8S_POD_NAMESPACE=%v", podEnv.K8S_POD_NAME, podEnv.K8S_POD_NAMESPACE)
	args.IfName = "azure0"

	// Create test data to delete
	_, addr, err := net.ParseCIDR("192.168.0.1/24")
	fmt.Println(err)
	epInfo := &network.EndpointInfo{
		IPAddresses: []net.IPNet{*addr},
	}
	err = plugin.nm.CreateEndpoint(nwCfg.Name, epInfo)
	/*
		if err != nil {
			t.Fatal(err)
		}*/

	nwInfo := &network.NetworkInfo{
		Id:      "test-nwcfg",
		Options: make(map[string]interface{}),
	}

	err = plugin.nm.CreateNetwork(nwInfo)
	/*
		if err != nil {
			t.Fatal(err)
		}*/

	//plugin.Add(args)
}
