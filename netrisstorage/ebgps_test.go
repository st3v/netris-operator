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

	"github.com/netrisai/netriswebapi/v2/types/bgp"
)

func TestNewBGPStorage(t *testing.T) {
	storage := NewBGPStorage()
	if storage == nil {
		t.Fatal("expected non-nil BGPStorage")
	}
	// NewBGPStorage initializes BGPs to an empty slice
	if storage.BGPs == nil {
		t.Error("expected BGPs to be initialized")
	}
	if len(storage.BGPs) != 0 {
		t.Errorf("expected empty BGPs slice, got %d items", len(storage.BGPs))
	}
}

func TestBGPStorage_GetAll_Empty(t *testing.T) {
	storage := NewBGPStorage()
	result := storage.GetAll()
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestBGPStorage_GetAll_WithData(t *testing.T) {
	storage := NewBGPStorage()
	storage.BGPs = []*bgp.EBGP{
		{ID: 1, Name: "bgp1"},
		{ID: 2, Name: "bgp2"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 BGPs, got %d", len(result))
	}
}

func TestBGPStorage_FindByName_Found(t *testing.T) {
	storage := NewBGPStorage()
	storage.BGPs = []*bgp.EBGP{
		{ID: 1, Name: "peer-upstream"},
		{ID: 2, Name: "peer-downstream"},
	}

	result, ok := storage.FindByName("peer-downstream")
	if !ok {
		t.Fatal("expected to find BGP 'peer-downstream'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestBGPStorage_FindByName_NotFound(t *testing.T) {
	storage := NewBGPStorage()
	storage.BGPs = []*bgp.EBGP{
		{ID: 1, Name: "peer-upstream"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find BGP 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestBGPStorage_FindByID_Found(t *testing.T) {
	storage := NewBGPStorage()
	storage.BGPs = []*bgp.EBGP{
		{ID: 100, Name: "bgp100"},
		{ID: 200, Name: "bgp200"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find BGP with ID 200")
	}
	if result.Name != "bgp200" {
		t.Errorf("expected name 'bgp200', got %q", result.Name)
	}
}
