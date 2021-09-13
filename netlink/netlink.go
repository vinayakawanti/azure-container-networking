// Copyright 2017 Microsoft. All rights reserved.
// MIT License

//go:build linux
// +build linux

package netlink

import (
	"net"

	"golang.org/x/sys/unix"
)

type NetlinkInterface interface {
	AddLink(link Link) error
	DeleteLink(name string) error
	SetLinkName(name string, newName string) error
	SetLinkState(name string, up bool) error
	SetLinkMaster(name string, master string) error
	SetLinkNetNs(name string, fd uintptr) error
	SetLinkAddress(ifName string, hwAddress net.HardwareAddr) error
	SetLinkPromisc(ifName string, on bool) error
	SetLinkHairpin(bridgeName string, on bool) error
	AddOrRemoveStaticArp(mode int, name string, ipaddr net.IP, mac net.HardwareAddr, isProxy bool) error
	GetIpAddressFamily(ip net.IP) int
	AddIpAddress(ifName string, ipAddress net.IP, ipNet *net.IPNet) error
	DeleteIpAddress(ifName string, ipAddress net.IP, ipNet *net.IPNet) error
	GetIpRoute(filter *Route) ([]*Route, error)
	AddIpRoute(route *Route) error
	DeleteIpRoute(route *Route) error
	ResetSocket()
}

type Netlink struct{}

// Init initializes netlink module.
func init() {
	initEncoder()
}

// Echo sends a netlink echo request message.
// TODO do we need this function?
func Echo(text string) error {
	s, err := getSocket()
	if err != nil {
		return err
	}

	req := newRequest(unix.NLMSG_NOOP, unix.NLM_F_ECHO|unix.NLM_F_ACK)
	if req == nil {
		return unix.ENOMEM
	}

	req.addPayload(newAttributeString(0, text))

	return s.sendAndWaitForAck(req)
}
