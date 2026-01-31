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

	"github.com/netrisai/netriswebapi/v2/types/nat"
)

func TestNewNATStorage(t *testing.T) {
	storage := NewNATStorage()
	if storage == nil {
		t.Fatal("expected non-nil NATStorage")
	}
	if storage.NAT != nil {
		t.Errorf("expected NAT to be nil, got %v", storage.NAT)
	}
}

func TestNATStorage_GetAll_Empty(t *testing.T) {
	storage := NewNATStorage()
	result := storage.GetAll()
	if result != nil {
		t.Errorf("expected nil for empty storage, got %v", result)
	}
}

func TestNATStorage_GetAll_WithData(t *testing.T) {
	storage := NewNATStorage()
	storage.NAT = []*nat.NAT{
		{ID: 1, Name: "nat1"},
		{ID: 2, Name: "nat2"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 NATs, got %d", len(result))
	}
}

func TestNATStorage_FindByName_Found(t *testing.T) {
	storage := NewNATStorage()
	storage.NAT = []*nat.NAT{
		{ID: 1, Name: "snat-rule"},
		{ID: 2, Name: "dnat-rule"},
	}

	result, ok := storage.FindByName("dnat-rule")
	if !ok {
		t.Fatal("expected to find NAT 'dnat-rule'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestNATStorage_FindByName_NotFound(t *testing.T) {
	storage := NewNATStorage()
	storage.NAT = []*nat.NAT{
		{ID: 1, Name: "snat-rule"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find NAT 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestNATStorage_FindByID_Found(t *testing.T) {
	storage := NewNATStorage()
	storage.NAT = []*nat.NAT{
		{ID: 100, Name: "nat100"},
		{ID: 200, Name: "nat200"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find NAT with ID 200")
	}
	if result.Name != "nat200" {
		t.Errorf("expected name 'nat200', got %q", result.Name)
	}
}
