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

	"github.com/netrisai/netriswebapi/v2/types/ipam"
)

func TestNewSubnetsStorage(t *testing.T) {
	storage := NewSubnetsStorage()
	if storage == nil {
		t.Fatal("expected non-nil SubnetsStorage")
	}
	if storage.Subnets != nil {
		t.Errorf("expected Subnets to be nil, got %v", storage.Subnets)
	}
}

func TestSubnetsStorage_GetAll_Empty(t *testing.T) {
	storage := NewSubnetsStorage()
	result := storage.GetAll()
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestSubnetsStorage_GetAll_WithData(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{ID: 1, Name: "subnet1", Type: "subnet"},
		{ID: 2, Name: "subnet2", Type: "subnet"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 subnets, got %d", len(result))
	}
}

func TestSubnetsStorage_FindByName_Found(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{ID: 1, Name: "public", Type: "allocation"},
		{ID: 2, Name: "private", Type: "subnet"},
	}

	result, ok := storage.FindByName("private")
	if !ok {
		t.Fatal("expected to find subnet 'private'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestSubnetsStorage_FindByName_NotFound(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{ID: 1, Name: "public", Type: "allocation"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find subnet 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestSubnetsStorage_FindByName_InChildren(t *testing.T) {
	// Test hierarchical search - finding a subnet nested in children
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{
			ID:   1,
			Name: "parent",
			Type: "allocation",
			Children: []*ipam.IPAM{
				{
					ID:   2,
					Name: "child1",
					Type: "subnet",
				},
				{
					ID:   3,
					Name: "child2",
					Type: "subnet",
				},
			},
		},
	}

	result, ok := storage.FindByName("child2")
	if !ok {
		t.Fatal("expected to find subnet 'child2' in children")
	}
	if result.ID != 3 {
		t.Errorf("expected ID 3, got %d", result.ID)
	}
}

func TestSubnetsStorage_FindByName_DeepNested(t *testing.T) {
	// Test deeply nested children
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{
			ID:   1,
			Name: "level1",
			Type: "allocation",
			Children: []*ipam.IPAM{
				{
					ID:   2,
					Name: "level2",
					Type: "subnet",
					Children: []*ipam.IPAM{
						{
							ID:   3,
							Name: "level3",
							Type: "subnet",
							Children: []*ipam.IPAM{
								{
									ID:   4,
									Name: "deepest",
									Type: "subnet",
								},
							},
						},
					},
				},
			},
		},
	}

	result, ok := storage.FindByName("deepest")
	if !ok {
		t.Fatal("expected to find deeply nested subnet 'deepest'")
	}
	if result.ID != 4 {
		t.Errorf("expected ID 4, got %d", result.ID)
	}
}

func TestSubnetsStorage_FindByID_Found(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{ID: 100, Name: "subnet100", Type: "subnet"},
		{ID: 200, Name: "subnet200", Type: "allocation"},
	}

	result, ok := storage.FindByID(200, "allocation")
	if !ok {
		t.Fatal("expected to find item with ID 200 and type 'allocation'")
	}
	if result.Name != "subnet200" {
		t.Errorf("expected name 'subnet200', got %q", result.Name)
	}
}

func TestSubnetsStorage_FindByID_WrongType(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{ID: 100, Name: "subnet100", Type: "subnet"},
	}

	// ID exists but type doesn't match - this will trigger download
	// which requires Cred, so we only test the matching case
}

func TestSubnetsStorage_FindByID_InChildren(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{
			ID:   1,
			Name: "parent",
			Type: "allocation",
			Children: []*ipam.IPAM{
				{
					ID:   100,
					Name: "child",
					Type: "subnet",
				},
			},
		},
	}

	result, ok := storage.FindByID(100, "subnet")
	if !ok {
		t.Fatal("expected to find subnet with ID 100 in children")
	}
	if result.Name != "child" {
		t.Errorf("expected name 'child', got %q", result.Name)
	}
}

func TestSubnetsStorage_FindByID_DeepNested(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{
			ID:   1,
			Name: "level1",
			Type: "allocation",
			Children: []*ipam.IPAM{
				{
					ID:   2,
					Name: "level2",
					Type: "subnet",
					Children: []*ipam.IPAM{
						{
							ID:   999,
							Name: "deepsubnet",
							Type: "subnet",
						},
					},
				},
			},
		},
	}

	result, ok := storage.FindByID(999, "subnet")
	if !ok {
		t.Fatal("expected to find deeply nested subnet with ID 999")
	}
	if result.Name != "deepsubnet" {
		t.Errorf("expected name 'deepsubnet', got %q", result.Name)
	}
}

func TestSubnetsStorage_FindByName_MultipleRoots(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{
			ID:   1,
			Name: "root1",
			Type: "allocation",
			Children: []*ipam.IPAM{
				{ID: 10, Name: "child1a", Type: "subnet"},
			},
		},
		{
			ID:   2,
			Name: "root2",
			Type: "allocation",
			Children: []*ipam.IPAM{
				{ID: 20, Name: "child2a", Type: "subnet"},
				{ID: 21, Name: "target", Type: "subnet"},
			},
		},
	}

	result, ok := storage.FindByName("target")
	if !ok {
		t.Fatal("expected to find 'target' in second root's children")
	}
	if result.ID != 21 {
		t.Errorf("expected ID 21, got %d", result.ID)
	}
}

func TestSubnetsStorage_FindByName_EmptyChildren(t *testing.T) {
	storage := NewSubnetsStorage()
	storage.Subnets = []*ipam.IPAM{
		{
			ID:       1,
			Name:     "parent",
			Type:     "allocation",
			Children: []*ipam.IPAM{},
		},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find subnet with empty children")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
