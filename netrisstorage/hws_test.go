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

	"github.com/netrisai/netriswebapi/v2/types/inventory"
)

func TestNewHWsStorage(t *testing.T) {
	storage := NewHWsStorage()
	if storage == nil {
		t.Fatal("expected non-nil HWsStorage")
	}
	if storage.HWs != nil {
		t.Errorf("expected HWs to be nil, got %v", storage.HWs)
	}
}

func TestHWsStorage_GetAll_Empty(t *testing.T) {
	storage := NewHWsStorage()
	result := storage.GetAll()
	if result != nil {
		t.Errorf("expected nil for empty storage, got %v", result)
	}
}

func TestHWsStorage_GetAll_WithData(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "switch1", Type: "switch"},
		{ID: 2, Name: "softgate1", Type: "softgate"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 HWs, got %d", len(result))
	}
}

func TestHWsStorage_FindByName_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "switch1", Type: "switch"},
		{ID: 2, Name: "softgate1", Type: "softgate"},
	}

	result, ok := storage.FindByName("softgate1")
	if !ok {
		t.Fatal("expected to find HW 'softgate1'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestHWsStorage_FindByName_NotFound(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "switch1", Type: "switch"},
	}

	result, ok := storage.FindByName("nonexistent")
	if ok {
		t.Error("expected not to find HW 'nonexistent'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestHWsStorage_FindSoftgateByName_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "sg1", Type: "softgate"},
		{ID: 2, Name: "sg1", Type: "switch"}, // Same name but different type
		{ID: 3, Name: "sg2", Type: "softgate"},
	}

	result, ok := storage.FindSoftgateByName("sg1")
	if !ok {
		t.Fatal("expected to find softgate 'sg1'")
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1 (softgate), got %d", result.ID)
	}
	if result.Type != "softgate" {
		t.Errorf("expected type 'softgate', got %q", result.Type)
	}
}

func TestHWsStorage_FindSoftgateByName_NotSoftgate(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "device1", Type: "switch"},
	}

	result, ok := storage.FindSoftgateByName("device1")
	if ok {
		t.Error("expected not to find softgate for switch type")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestHWsStorage_FindSwitchByName_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "sw1", Type: "switch"},
		{ID: 2, Name: "sw1", Type: "softgate"}, // Same name but different type
	}

	result, ok := storage.FindSwitchByName("sw1")
	if !ok {
		t.Fatal("expected to find switch 'sw1'")
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1 (switch), got %d", result.ID)
	}
	if result.Type != "switch" {
		t.Errorf("expected type 'switch', got %q", result.Type)
	}
}

func TestHWsStorage_FindSwitchByName_NotSwitch(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "device1", Type: "softgate"},
	}

	result, ok := storage.FindSwitchByName("device1")
	if ok {
		t.Error("expected not to find switch for softgate type")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestHWsStorage_FindControllerByName_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "ctrl1", Type: "controller"},
		{ID: 2, Name: "ctrl1", Type: "switch"},
	}

	result, ok := storage.FindControllerByName("ctrl1")
	if !ok {
		t.Fatal("expected to find controller 'ctrl1'")
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1 (controller), got %d", result.ID)
	}
	if result.Type != "controller" {
		t.Errorf("expected type 'controller', got %q", result.Type)
	}
}

func TestHWsStorage_FindControllerByName_NotController(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "device1", Type: "switch"},
	}

	result, ok := storage.FindControllerByName("device1")
	if ok {
		t.Error("expected not to find controller for switch type")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestHWsStorage_FindByID_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 100, Name: "hw100", Type: "switch"},
		{ID: 200, Name: "hw200", Type: "softgate"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find HW with ID 200")
	}
	if result.Name != "hw200" {
		t.Errorf("expected name 'hw200', got %q", result.Name)
	}
}

func TestHWsStorage_FindSoftgateByID_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 100, Name: "hw100", Type: "switch"},
		{ID: 200, Name: "hw200", Type: "softgate"},
	}

	result, ok := storage.FindSoftgateByID(200)
	if !ok {
		t.Fatal("expected to find softgate with ID 200")
	}
	if result.Type != "softgate" {
		t.Errorf("expected type 'softgate', got %q", result.Type)
	}
}


func TestHWsStorage_FindSwitchByID_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 100, Name: "hw100", Type: "switch"},
		{ID: 200, Name: "hw200", Type: "softgate"},
	}

	result, ok := storage.FindSwitchByID(100)
	if !ok {
		t.Fatal("expected to find switch with ID 100")
	}
	if result.Type != "switch" {
		t.Errorf("expected type 'switch', got %q", result.Type)
	}
}

func TestHWsStorage_FindControllerByID_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 100, Name: "hw100", Type: "controller"},
		{ID: 200, Name: "hw200", Type: "softgate"},
	}

	result, ok := storage.FindControllerByID(100)
	if !ok {
		t.Fatal("expected to find controller with ID 100")
	}
	if result.Type != "controller" {
		t.Errorf("expected type 'controller', got %q", result.Type)
	}
}

func TestHWsStorage_FindHWsBySite_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "sw1", Type: "switch", Site: inventory.IDName{ID: 10, Name: "site1"}},
		{ID: 2, Name: "sw2", Type: "switch", Site: inventory.IDName{ID: 20, Name: "site2"}},
		{ID: 3, Name: "sg1", Type: "softgate", Site: inventory.IDName{ID: 10, Name: "site1"}},
	}

	result := storage.FindHWsBySite(10)
	if len(result) != 2 {
		t.Errorf("expected 2 HWs in site 10, got %d", len(result))
	}
}

func TestHWsStorage_FindHWsBySite_Empty(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "sw1", Type: "switch", Site: inventory.IDName{ID: 10, Name: "site1"}},
	}

	result := storage.FindHWsBySite(99)
	if len(result) != 0 {
		t.Errorf("expected 0 HWs in site 99, got %d", len(result))
	}
}

func TestHWsStorage_FindSpineBySite_Found(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "sw1", Type: "switch", Site: inventory.IDName{ID: 10, Name: "site1"}},
		{ID: 2, Name: "spine1", Type: "spine", Site: inventory.IDName{ID: 10, Name: "site1"}},
		{ID: 3, Name: "spine2", Type: "spine", Site: inventory.IDName{ID: 20, Name: "site2"}},
	}

	result := storage.FindSpineBySite(10)
	if result == nil {
		t.Fatal("expected to find spine in site 10")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
	if result.Type != "spine" {
		t.Errorf("expected type 'spine', got %q", result.Type)
	}
}

func TestHWsStorage_FindSpineBySite_NotFound(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "sw1", Type: "switch", Site: inventory.IDName{ID: 10, Name: "site1"}},
	}

	result := storage.FindSpineBySite(10)
	if result != nil {
		t.Errorf("expected nil (no spine in site), got %v", result)
	}
}

func TestHWsStorage_FindSpineBySite_WrongSite(t *testing.T) {
	storage := NewHWsStorage()
	storage.HWs = []*inventory.HW{
		{ID: 1, Name: "spine1", Type: "spine", Site: inventory.IDName{ID: 10, Name: "site1"}},
	}

	result := storage.FindSpineBySite(99)
	if result != nil {
		t.Errorf("expected nil (spine in different site), got %v", result)
	}
}
