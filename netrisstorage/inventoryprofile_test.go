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

	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"
)

func TestNewInventoryProfileStorage(t *testing.T) {
	storage := NewInventoryProfileStorage()
	if storage == nil {
		t.Fatal("expected non-nil InventoryProfileStorage")
	}
	if storage.InventoryProfile != nil {
		t.Errorf("expected InventoryProfile to be nil, got %v", storage.InventoryProfile)
	}
}

func TestInventoryProfileStorage_GetAll_Empty(t *testing.T) {
	storage := NewInventoryProfileStorage()
	result := storage.GetAll()
	if result != nil {
		t.Errorf("expected nil for empty storage, got %v", result)
	}
}

func TestInventoryProfileStorage_GetAll_WithData(t *testing.T) {
	storage := NewInventoryProfileStorage()
	storage.InventoryProfile = []*inventoryprofile.Profile{
		{ID: 1, Name: "profile1"},
		{ID: 2, Name: "profile2"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(result))
	}
}

func TestInventoryProfileStorage_FindByName_Found(t *testing.T) {
	storage := NewInventoryProfileStorage()
	storage.InventoryProfile = []*inventoryprofile.Profile{
		{ID: 1, Name: "default"},
		{ID: 2, Name: "custom"},
	}

	result, ok := storage.FindByName("custom")
	if !ok {
		t.Fatal("expected to find profile 'custom'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestInventoryProfileStorage_FindByName_NotFound(t *testing.T) {
	storage := NewInventoryProfileStorage()
	storage.InventoryProfile = []*inventoryprofile.Profile{
		{ID: 1, Name: "default"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find profile 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestInventoryProfileStorage_FindByID_Found(t *testing.T) {
	storage := NewInventoryProfileStorage()
	storage.InventoryProfile = []*inventoryprofile.Profile{
		{ID: 100, Name: "profile100"},
		{ID: 200, Name: "profile200"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find profile with ID 200")
	}
	if result.Name != "profile200" {
		t.Errorf("expected name 'profile200', got %q", result.Name)
	}
}
