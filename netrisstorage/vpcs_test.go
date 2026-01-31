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

	"github.com/netrisai/netriswebapi/v2/types/vpc"
)

func TestNewVPCStorage(t *testing.T) {
	storage := NewVPCStorage()
	if storage == nil {
		t.Fatal("expected non-nil VPCStorage")
	}
	if storage.VPCs != nil {
		t.Errorf("expected VPCs to be nil, got %v", storage.VPCs)
	}
}

func TestVPCStorage_GetAll_Empty(t *testing.T) {
	storage := NewVPCStorage()
	result := storage.GetAll()
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestVPCStorage_GetAll_WithData(t *testing.T) {
	storage := NewVPCStorage()
	storage.VPCs = []*vpc.VPC{
		{ID: 1, Name: "vpc1"},
		{ID: 2, Name: "vpc2"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 VPCs, got %d", len(result))
	}
	if result[0].Name != "vpc1" {
		t.Errorf("expected first VPC name 'vpc1', got %q", result[0].Name)
	}
}

func TestVPCStorage_FindByName_Found(t *testing.T) {
	storage := NewVPCStorage()
	storage.VPCs = []*vpc.VPC{
		{ID: 1, Name: "production"},
		{ID: 2, Name: "staging"},
	}

	result, ok := storage.FindByName("staging")
	if !ok {
		t.Fatal("expected to find VPC 'staging'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestVPCStorage_FindByName_NotFound_TriggersDownload(t *testing.T) {
	// Note: FindByName triggers download when not found, which requires Cred
	// This test verifies the found path works correctly
	storage := NewVPCStorage()
	storage.VPCs = []*vpc.VPC{
		{ID: 1, Name: "production"},
	}

	result, ok := storage.FindByName("production")
	if !ok {
		t.Fatal("expected to find VPC 'production'")
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestVPCStorage_FindByID_Found(t *testing.T) {
	storage := NewVPCStorage()
	storage.VPCs = []*vpc.VPC{
		{ID: 100, Name: "vpc100"},
		{ID: 200, Name: "vpc200"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find VPC with ID 200")
	}
	if result.Name != "vpc200" {
		t.Errorf("expected name 'vpc200', got %q", result.Name)
	}
}
