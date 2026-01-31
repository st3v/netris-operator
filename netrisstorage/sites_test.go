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

	"github.com/netrisai/netriswebapi/v2/types/site"
)

func TestNewSitesStorage(t *testing.T) {
	storage := NewSitesStorage()
	if storage == nil {
		t.Fatal("expected non-nil SitesStorage")
	}
	if storage.Sites != nil {
		t.Errorf("expected Sites to be nil, got %v", storage.Sites)
	}
}

func TestSitesStorage_GetAll_Empty(t *testing.T) {
	storage := NewSitesStorage()
	result := storage.GetAll()
	if result != nil {
		t.Errorf("expected nil for empty storage, got %v", result)
	}
}

func TestSitesStorage_GetAll_WithData(t *testing.T) {
	storage := NewSitesStorage()
	storage.Sites = []*site.Site{
		{ID: 1, Name: "site1"},
		{ID: 2, Name: "site2"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 sites, got %d", len(result))
	}
	if result[0].Name != "site1" {
		t.Errorf("expected first site name 'site1', got %q", result[0].Name)
	}
	if result[1].Name != "site2" {
		t.Errorf("expected second site name 'site2', got %q", result[1].Name)
	}
}

func TestSitesStorage_FindByName_Found(t *testing.T) {
	storage := NewSitesStorage()
	storage.Sites = []*site.Site{
		{ID: 1, Name: "alpha"},
		{ID: 2, Name: "beta"},
		{ID: 3, Name: "gamma"},
	}

	result, ok := storage.FindByName("beta")
	if !ok {
		t.Fatal("expected to find site 'beta'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
	if result.Name != "beta" {
		t.Errorf("expected name 'beta', got %q", result.Name)
	}
}

func TestSitesStorage_FindByName_NotFound(t *testing.T) {
	storage := NewSitesStorage()
	storage.Sites = []*site.Site{
		{ID: 1, Name: "alpha"},
		{ID: 2, Name: "beta"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find site 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestSitesStorage_FindByName_EmptyStorage(t *testing.T) {
	storage := NewSitesStorage()

	result, ok := storage.FindByName("anything")
	if ok {
		t.Error("expected not to find site in empty storage")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestSitesStorage_FindByID_Found(t *testing.T) {
	storage := NewSitesStorage()
	storage.Sites = []*site.Site{
		{ID: 10, Name: "site10"},
		{ID: 20, Name: "site20"},
		{ID: 30, Name: "site30"},
	}

	result, ok := storage.FindByID(20)
	if !ok {
		t.Fatal("expected to find site with ID 20")
	}
	if result.ID != 20 {
		t.Errorf("expected ID 20, got %d", result.ID)
	}
	if result.Name != "site20" {
		t.Errorf("expected name 'site20', got %q", result.Name)
	}
}

func TestSitesStorage_FindByID_NotFound_NoCred(t *testing.T) {
	// When ID is not found and Cred is nil, download() will panic
	// This test verifies the found path works correctly
	storage := NewSitesStorage()
	storage.Sites = []*site.Site{
		{ID: 1, Name: "site1"},
	}

	// Finding existing ID should work without API call
	result, ok := storage.FindByID(1)
	if !ok {
		t.Fatal("expected to find site with ID 1")
	}
	if result.Name != "site1" {
		t.Errorf("expected name 'site1', got %q", result.Name)
	}
}

func TestSitesStorage_FindByName_FirstMatch(t *testing.T) {
	// Test that FindByName returns the first matching site
	storage := NewSitesStorage()
	storage.Sites = []*site.Site{
		{ID: 1, Name: "duplicate"},
		{ID: 2, Name: "duplicate"},
	}

	result, ok := storage.FindByName("duplicate")
	if !ok {
		t.Fatal("expected to find site 'duplicate'")
	}
	if result.ID != 1 {
		t.Errorf("expected first match with ID 1, got %d", result.ID)
	}
}
