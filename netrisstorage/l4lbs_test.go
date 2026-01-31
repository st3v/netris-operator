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

	"github.com/netrisai/netriswebapi/v2/types/l4lb"
)

func TestNewL4LBStorage(t *testing.T) {
	storage := NewL4LBStorage()
	if storage == nil {
		t.Fatal("expected non-nil L4LBStorage")
	}
	if storage.L4LBs != nil {
		t.Errorf("expected L4LBs to be nil, got %v", storage.L4LBs)
	}
}

func TestL4LBStorage_GetAll_Empty(t *testing.T) {
	storage := NewL4LBStorage()
	result := storage.GetAll()
	if result != nil {
		t.Errorf("expected nil for empty storage, got %v", result)
	}
}

func TestL4LBStorage_GetAll_WithData(t *testing.T) {
	storage := NewL4LBStorage()
	storage.L4LBs = []*l4lb.LoadBalancer{
		{ID: 1, Name: "lb1"},
		{ID: 2, Name: "lb2"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 L4LBs, got %d", len(result))
	}
	if result[0].Name != "lb1" {
		t.Errorf("expected first LB name 'lb1', got %q", result[0].Name)
	}
}

func TestL4LBStorage_FindByName_Found(t *testing.T) {
	storage := NewL4LBStorage()
	storage.L4LBs = []*l4lb.LoadBalancer{
		{ID: 1, Name: "web-lb"},
		{ID: 2, Name: "api-lb"},
		{ID: 3, Name: "db-lb"},
	}

	result, ok := storage.FindByName("api-lb")
	if !ok {
		t.Fatal("expected to find L4LB 'api-lb'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
	if result.Name != "api-lb" {
		t.Errorf("expected name 'api-lb', got %q", result.Name)
	}
}

func TestL4LBStorage_FindByName_NotFound(t *testing.T) {
	storage := NewL4LBStorage()
	storage.L4LBs = []*l4lb.LoadBalancer{
		{ID: 1, Name: "web-lb"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find L4LB 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestL4LBStorage_FindByName_EmptyStorage(t *testing.T) {
	storage := NewL4LBStorage()

	result, ok := storage.FindByName("anything")
	if ok {
		t.Error("expected not to find L4LB in empty storage")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestL4LBStorage_FindByID_Found(t *testing.T) {
	storage := NewL4LBStorage()
	storage.L4LBs = []*l4lb.LoadBalancer{
		{ID: 100, Name: "lb100"},
		{ID: 200, Name: "lb200"},
		{ID: 300, Name: "lb300"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find L4LB with ID 200")
	}
	if result.Name != "lb200" {
		t.Errorf("expected name 'lb200', got %q", result.Name)
	}
}

func TestL4LBStorage_FindByName_FirstMatch(t *testing.T) {
	// Test that FindByName returns the first matching L4LB
	storage := NewL4LBStorage()
	storage.L4LBs = []*l4lb.LoadBalancer{
		{ID: 1, Name: "duplicate"},
		{ID: 2, Name: "duplicate"},
	}

	result, ok := storage.FindByName("duplicate")
	if !ok {
		t.Fatal("expected to find L4LB 'duplicate'")
	}
	if result.ID != 1 {
		t.Errorf("expected first match with ID 1, got %d", result.ID)
	}
}
