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

	"github.com/netrisai/netriswebapi/v1/types/tenant"
)

func TestNewTenantsStorage(t *testing.T) {
	storage := NewTenantsStorage()
	if storage == nil {
		t.Fatal("expected non-nil TenantsStorage")
	}
	if storage.Tenants != nil {
		t.Errorf("expected Tenants to be nil, got %v", storage.Tenants)
	}
}

func TestTenantsStorage_GetAll_Empty(t *testing.T) {
	storage := NewTenantsStorage()
	result := storage.GetAll()
	if result != nil {
		t.Errorf("expected nil for empty storage, got %v", result)
	}
}

func TestTenantsStorage_GetAll_WithData(t *testing.T) {
	storage := NewTenantsStorage()
	storage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "dev"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(result))
	}
	if result[0].Name != "admin" {
		t.Errorf("expected first tenant name 'admin', got %q", result[0].Name)
	}
}

func TestTenantsStorage_FindByName_Found(t *testing.T) {
	storage := NewTenantsStorage()
	storage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "development"},
		{ID: 3, Name: "production"},
	}

	result, ok := storage.FindByName("development")
	if !ok {
		t.Fatal("expected to find tenant 'development'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestTenantsStorage_FindByName_NotFound(t *testing.T) {
	storage := NewTenantsStorage()
	storage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find tenant 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestTenantsStorage_FindByName_EmptyStorage(t *testing.T) {
	storage := NewTenantsStorage()

	result, ok := storage.FindByName("anything")
	if ok {
		t.Error("expected not to find tenant in empty storage")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestTenantsStorage_FindByID_Found(t *testing.T) {
	storage := NewTenantsStorage()
	storage.Tenants = []*tenant.Tenant{
		{ID: 100, Name: "tenant100"},
		{ID: 200, Name: "tenant200"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find tenant with ID 200")
	}
	if result.Name != "tenant200" {
		t.Errorf("expected name 'tenant200', got %q", result.Name)
	}
}
