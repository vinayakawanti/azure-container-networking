// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package imdsclient

import (
	"encoding/xml"
)

const (
	hostQueryURL                     = "http://168.63.129.16/machine/plugins?comp=nmagent&type=getinterfaceinfov1"
	hostQueryURLForProgrammedVersion = "http://168.63.129.16/machine/plugins/?comp=nmagent&type=NetworkManagement/interfaces/%s/networkContainers/%s/authenticationToken/%s/api-version/%s"
)

// ImdsClient can be used to connect to VM Host agent in Azure.
type ImdsClient struct {
	primaryInterface *InterfaceInfo
}

// InterfaceInfo specifies the information about an interface as returned by Host Agent.
type InterfaceInfo struct {
	Subnet       string
	Gateway      string
	IsPrimary    bool
	PrimaryIP    string
	SecondaryIPs []string
}

// Azure host agent XML document format.
type xmlDocument struct {
	XMLName   xml.Name `xml:"Interfaces"`
	Interface []struct {
		XMLName    xml.Name `xml:"Interface"`
		MacAddress string   `xml:"MacAddress,attr"`
		IsPrimary  bool     `xml:"IsPrimary,attr"`

		IPSubnet []struct {
			XMLName xml.Name `xml:"IPSubnet"`
			Prefix  string   `xml:"Prefix,attr"`

			IPAddress []struct {
				XMLName   xml.Name `xml:"IPAddress"`
				Address   string   `xml:"Address,attr"`
				IsPrimary bool     `xml:"IsPrimary,attr"`
			}
		}
	}
}

type containerVersionJsonResponse struct {
	HTTPResponseCode   string `json:"httpResponseCode"`
	NetworkContainerID string `json:"networkContainerId"`
	ProgrammedVersion  string `json:"Version"`
}

// InterfaceInfo specifies the information about an interface as returned by Host Agent.
type ContainerVersion struct {
	NetworkContainerID string
	ProgrammedVersion  string
}

// An ImdsInterface performs CRUD operations on IP reservations
type ImdsClientInterface interface {
	GetNetworkContainerInfoFromHost(networkContainerID string, primaryAddress string, authToken string, apiVersion string) (*ContainerVersion, error)
	GetPrimaryInterfaceInfoFromHost() (*InterfaceInfo, error)
	GetPrimaryInterfaceInfoFromMemory() (*InterfaceInfo, error)
	GetNetworkContainerInfoFromHostWithoutToken() int
}
