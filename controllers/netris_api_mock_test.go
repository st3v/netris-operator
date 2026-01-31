/*
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
	"encoding/json"

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

// successReply returns a successful HTTP reply for mock operations.
func successReply(id int) http.HTTPReply {
	data, _ := json.Marshal(map[string]interface{}{
		"isSuccess": true,
		"data":      map[string]int{"id": id},
	})
	return http.HTTPReply{Data: data, StatusCode: 200}
}

// errorReply returns an error HTTP reply for mock operations.
func errorReply(message string) http.HTTPReply {
	data, _ := json.Marshal(map[string]interface{}{
		"isSuccess": false,
		"message":   message,
	})
	return http.HTTPReply{Data: data}
}

// MockBGPClient implements BGPClient for testing.
type MockBGPClient struct {
	AddFunc    func(bgp *bgp.EBGPAdd) (http.HTTPReply, error)
	UpdateFunc func(id int, bgp *bgp.EBGPUpdate) (http.HTTPReply, error)
	DeleteFunc func(id int) (http.HTTPReply, error)
	AddErr     error
	UpdateErr  error
	DeleteErr  error
	LastAddID  int
}

func (m *MockBGPClient) Add(b *bgp.EBGPAdd) (http.HTTPReply, error) {
	if m.AddFunc != nil {
		return m.AddFunc(b)
	}
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockBGPClient) Update(id int, b *bgp.EBGPUpdate) (http.HTTPReply, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, b)
	}
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockBGPClient) Delete(id int) (http.HTTPReply, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}

// MockDHCPClient implements DHCPClient for testing.
type MockDHCPClient struct {
	Data   []*dhcp.DHCPOptionSet
	GetErr error
}

func (m *MockDHCPClient) Get() ([]*dhcp.DHCPOptionSet, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.Data, nil
}

// MockInventoryClient implements InventoryClient for testing.
type MockInventoryClient struct {
	HWData    []*inventory.HW
	NOSData   []*inventory.NOS
	GetErr    error
	GetNOSErr error
	AddErr    error
	UpdateErr error
	DeleteErr error
	LastAddID int
}

func (m *MockInventoryClient) Get() ([]*inventory.HW, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.HWData, nil
}

func (m *MockInventoryClient) GetNOS() ([]*inventory.NOS, error) {
	if m.GetNOSErr != nil {
		return nil, m.GetNOSErr
	}
	return m.NOSData, nil
}

func (m *MockInventoryClient) AddController(hw *inventory.HWController) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockInventoryClient) AddSoftgate(hw *inventory.HWSoftgate) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockInventoryClient) AddSwitch(hw *inventory.HWSwitchAdd) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockInventoryClient) UpdateController(id int, hw *inventory.HWControllerUpdate) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockInventoryClient) UpdateSoftgate(id int, hw *inventory.HWSoftgateUpdate) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockInventoryClient) UpdateSwitch(id int, hw *inventory.HWSwitchUpdate) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockInventoryClient) Delete(kind string, id int) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}

// MockInventoryProfileClient implements InventoryProfileClient for testing.
type MockInventoryProfileClient struct {
	Data      []*inventoryprofile.Profile
	GetErr    error
	AddErr    error
	UpdateErr error
	DeleteErr error
	LastAddID int
}

func (m *MockInventoryProfileClient) Get() ([]*inventoryprofile.Profile, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.Data, nil
}

func (m *MockInventoryProfileClient) Add(profile *inventoryprofile.ProfileW) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockInventoryProfileClient) Update(profile *inventoryprofile.ProfileW) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(profile.ID), nil
}

func (m *MockInventoryProfileClient) Delete(id int) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}

// MockIPAMClient implements IPAMClient for testing.
type MockIPAMClient struct {
	Data      []*ipam.IPAM
	GetErr    error
	AddErr    error
	UpdateErr error
	DeleteErr error
	LastAddID int
}

func (m *MockIPAMClient) Get() ([]*ipam.IPAM, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.Data, nil
}

func (m *MockIPAMClient) AddAllocation(allocation *ipam.Allocation) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockIPAMClient) AddSubnet(subnet *ipam.Subnet) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockIPAMClient) UpdateAllocation(id int, allocation *ipam.Allocation) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockIPAMClient) UpdateSubnet(id int, subnet *ipam.Subnet) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockIPAMClient) Delete(kind string, id int) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}

// MockL4LBClient implements L4LBClient for testing.
type MockL4LBClient struct {
	AddErr    error
	UpdateErr error
	DeleteErr error
	LastAddID int
}

func (m *MockL4LBClient) Add(lb *l4lb.LoadBalancerAdd) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockL4LBClient) Update(id int, lb *l4lb.LoadBalancerUpdate) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockL4LBClient) Delete(id int) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}

// MockLinkClient implements LinkClient for testing.
type MockLinkClient struct {
	AddErr    error
	DeleteErr error
	LastAddID int
}

func (m *MockLinkClient) Add(lnk *link.Linkw) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockLinkClient) Delete(lnk *link.Link) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(lnk.ID), nil
}

// MockNATClient implements NATClient for testing.
type MockNATClient struct {
	AddErr    error
	UpdateErr error
	DeleteErr error
	LastAddID int
}

func (m *MockNATClient) Add(n *nat.NATw) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockNATClient) Update(id int, n *nat.NATw) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockNATClient) Delete(id int) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}

// MockSiteClient implements SiteClient for testing.
type MockSiteClient struct {
	AddErr    error
	UpdateErr error
	DeleteErr error
	LastAddID int
}

func (m *MockSiteClient) Add(s *site.Site) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockSiteClient) Update(id int, s *site.Site) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockSiteClient) Delete(id int) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}

// MockVNetClient implements VNetClient for testing.
type MockVNetClient struct {
	Data         []*vnet.VNet
	DetailedData map[int]*vnet.VNetDetailed
	GetErr       error
	GetByIDErr   error
	AddErr       error
	UpdateErr    error
	DeleteErr    error
	LastAddID    int
}

func (m *MockVNetClient) Get() ([]*vnet.VNet, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.Data, nil
}

func (m *MockVNetClient) GetByID(id int) (*vnet.VNetDetailed, error) {
	if m.GetByIDErr != nil {
		return nil, m.GetByIDErr
	}
	if m.DetailedData != nil {
		if v, ok := m.DetailedData[id]; ok {
			return v, nil
		}
	}
	return nil, nil
}

func (m *MockVNetClient) Add(vn *vnet.VNetAdd) (http.HTTPReply, error) {
	if m.AddErr != nil {
		return http.HTTPReply{}, m.AddErr
	}
	m.LastAddID++
	return successReply(m.LastAddID), nil
}

func (m *MockVNetClient) Update(id int, vn *vnet.VNetUpdate) (http.HTTPReply, error) {
	if m.UpdateErr != nil {
		return http.HTTPReply{}, m.UpdateErr
	}
	return successReply(id), nil
}

func (m *MockVNetClient) Delete(id int) (http.HTTPReply, error) {
	if m.DeleteErr != nil {
		return http.HTTPReply{}, m.DeleteErr
	}
	return successReply(id), nil
}
