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

package netrisstorage

import (
	"testing"

	"github.com/netrisai/netriswebapi/v2/types/vnet"
)

func TestNewVNetStorage(t *testing.T) {
	storage := NewVNetStorage()
	if storage == nil {
		t.Fatal("expected non-nil VNetStorage")
	}
	if storage.VNets != nil {
		t.Errorf("expected VNets to be nil, got %v", storage.VNets)
	}
}

func TestVNetStorage_GetAll_Empty(t *testing.T) {
	storage := NewVNetStorage()
	result := storage.GetAll()
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestVNetStorage_GetAll_WithData(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{ID: 1, Name: "vnet1"},
		{ID: 2, Name: "vnet2"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 vnets, got %d", len(result))
	}
	if result[0].Name != "vnet1" {
		t.Errorf("expected first vnet name 'vnet1', got %q", result[0].Name)
	}
}

func TestVNetStorage_FindByName_Found(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{ID: 1, Name: "frontend"},
		{ID: 2, Name: "backend"},
		{ID: 3, Name: "database"},
	}

	result, ok := storage.FindByName("backend")
	if !ok {
		t.Fatal("expected to find vnet 'backend'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestVNetStorage_FindByName_NotFound(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{ID: 1, Name: "frontend"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find vnet 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestVNetStorage_FindByID_Found(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{ID: 100, Name: "vnet100"},
		{ID: 200, Name: "vnet200"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find vnet with ID 200")
	}
	if result.Name != "vnet200" {
		t.Errorf("expected name 'vnet200', got %q", result.Name)
	}
}

func TestVNetStorage_FindByGateway_ExactMatch(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{
			ID:   1,
			Name: "vnet1",
			Gateways: []vnet.VNetGateway{
				{Prefix: "10.0.0.1/24"},
			},
		},
		{
			ID:   2,
			Name: "vnet2",
			Gateways: []vnet.VNetGateway{
				{Prefix: "192.168.1.1/24"},
			},
		},
	}

	result, ok := storage.FindByGateway("10.0.0.1/24")
	if !ok {
		t.Fatal("expected to find vnet by gateway '10.0.0.1/24'")
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestVNetStorage_FindByGateway_CIDRContains(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{
			ID:   1,
			Name: "vnet1",
			Gateways: []vnet.VNetGateway{
				{Prefix: "10.0.0.5/24"},
			},
		},
	}

	// Search with CIDR that contains the gateway IP
	result, ok := storage.FindByGateway("10.0.0.0/24")
	if !ok {
		t.Fatal("expected to find vnet with gateway contained in 10.0.0.0/24")
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestVNetStorage_FindByGateway_NotFound(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{
			ID:   1,
			Name: "vnet1",
			Gateways: []vnet.VNetGateway{
				{Prefix: "10.0.0.1/24"},
			},
		},
	}

	result, ok := storage.FindByGateway("192.168.1.0/24")
	if ok {
		t.Error("expected not to find vnet with gateway '192.168.1.0/24'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestVNetStorage_FindByGateway_InvalidCIDR(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{
			ID:   1,
			Name: "vnet1",
			Gateways: []vnet.VNetGateway{
				{Prefix: "10.0.0.1/24"},
			},
		},
	}

	// Invalid CIDR should return not found
	result, ok := storage.FindByGateway("invalid-cidr")
	if ok {
		t.Error("expected not to find vnet with invalid CIDR")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestVNetStorage_FindByGateway_MultipleGateways(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{
			ID:   1,
			Name: "vnet1",
			Gateways: []vnet.VNetGateway{
				{Prefix: "10.0.0.1/24"},
				{Prefix: "10.0.1.1/24"},
				{Prefix: "192.168.1.1/24"},
			},
		},
	}

	// Should find by second gateway
	result, ok := storage.FindByGateway("10.0.1.1/24")
	if !ok {
		t.Fatal("expected to find vnet by second gateway")
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestVNetStorage_FindByGateway_EmptyGateways(t *testing.T) {
	storage := NewVNetStorage()
	storage.VNets = []*vnet.VNet{
		{
			ID:       1,
			Name:     "vnet1",
			Gateways: []vnet.VNetGateway{},
		},
	}

	result, ok := storage.FindByGateway("10.0.0.0/24")
	if ok {
		t.Error("expected not to find vnet with empty gateways")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
