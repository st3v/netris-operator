/*
Copyright 2021. Netris, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"
	"github.com/netrisai/netriswebapi/v2/types/bgp"
	"github.com/netrisai/netriswebapi/v2/types/dhcp"
	"github.com/netrisai/netriswebapi/v2/types/inventory"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/l4lb"
	"github.com/netrisai/netriswebapi/v2/types/link"
	"github.com/netrisai/netriswebapi/v2/types/nat"
	"github.com/netrisai/netriswebapi/v2/types/site"
	"github.com/netrisai/netriswebapi/v2/types/vnet"
)

// BGPClient defines the BGP operations used by controllers.
// Implemented by *bgp.BGPClient from netriswebapi.
type BGPClient interface {
	Add(bgp *bgp.EBGPAdd) (http.HTTPReply, error)
	Update(id int, bgp *bgp.EBGPUpdate) (http.HTTPReply, error)
	Delete(id int) (http.HTTPReply, error)
}

// DHCPClient defines the DHCP operations used by controllers.
// Implemented by *dhcp.Client from netriswebapi.
type DHCPClient interface {
	Get() ([]*dhcp.DHCPOptionSet, error)
}

// InventoryClient defines the Inventory operations used by controllers.
// Implemented by *inventory.Client from netriswebapi.
type InventoryClient interface {
	Get() ([]*inventory.HW, error)
	GetNOS() ([]*inventory.NOS, error)
	AddController(hw *inventory.HWController) (http.HTTPReply, error)
	AddSoftgate(hw *inventory.HWSoftgate) (http.HTTPReply, error)
	AddSwitch(hw *inventory.HWSwitchAdd) (http.HTTPReply, error)
	UpdateController(id int, hw *inventory.HWControllerUpdate) (http.HTTPReply, error)
	UpdateSoftgate(id int, hw *inventory.HWSoftgateUpdate) (http.HTTPReply, error)
	UpdateSwitch(id int, hw *inventory.HWSwitchUpdate) (http.HTTPReply, error)
	Delete(kind string, id int) (http.HTTPReply, error)
}

// InventoryProfileClient defines the InventoryProfile operations used by controllers.
// Implemented by *inventoryprofile.Client from netriswebapi.
type InventoryProfileClient interface {
	Get() ([]*inventoryprofile.Profile, error)
	Add(profile *inventoryprofile.ProfileW) (http.HTTPReply, error)
	Update(profile *inventoryprofile.ProfileW) (http.HTTPReply, error)
	Delete(id int) (http.HTTPReply, error)
}

// IPAMClient defines the IPAM operations used by controllers.
// Implemented by *ipam.IPAMClient from netriswebapi.
type IPAMClient interface {
	Get() ([]*ipam.IPAM, error)
	AddAllocation(allocation *ipam.Allocation) (http.HTTPReply, error)
	AddSubnet(subnet *ipam.Subnet) (http.HTTPReply, error)
	UpdateAllocation(id int, allocation *ipam.Allocation) (http.HTTPReply, error)
	UpdateSubnet(id int, subnet *ipam.Subnet) (http.HTTPReply, error)
	Delete(kind string, id int) (http.HTTPReply, error)
}

// L4LBClient defines the L4LB operations used by controllers.
// Implemented by *l4lb.LBClient from netriswebapi.
type L4LBClient interface {
	Add(l4lb *l4lb.LoadBalancerAdd) (http.HTTPReply, error)
	Update(id int, l4lb *l4lb.LoadBalancerUpdate) (http.HTTPReply, error)
	Delete(id int) (http.HTTPReply, error)
}

// LinkClient defines the Link operations used by controllers.
// Implemented by *link.Client from netriswebapi.
type LinkClient interface {
	Add(lnk *link.Linkw) (http.HTTPReply, error)
	Delete(lnk *link.Link) (http.HTTPReply, error)
}

// NATClient defines the NAT operations used by controllers.
// Implemented by *nat.Client from netriswebapi.
type NATClient interface {
	Add(nat *nat.NATw) (http.HTTPReply, error)
	Update(id int, nat *nat.NATw) (http.HTTPReply, error)
	Delete(id int) (http.HTTPReply, error)
}

// SiteClient defines the Site operations used by controllers.
// Implemented by *site.Client from netriswebapi.
type SiteClient interface {
	Add(site *site.Site) (http.HTTPReply, error)
	Update(id int, site *site.Site) (http.HTTPReply, error)
	Delete(id int) (http.HTTPReply, error)
}

// VNetClient defines the VNet operations used by controllers.
// Implemented by *vnet.VNetClient from netriswebapi.
type VNetClient interface {
	Get() ([]*vnet.VNet, error)
	GetByID(id int) (*vnet.VNetDetailed, error)
	Add(vn *vnet.VNetAdd) (http.HTTPReply, error)
	Update(id int, vn *vnet.VNetUpdate) (http.HTTPReply, error)
	Delete(id int) (http.HTTPReply, error)
}
