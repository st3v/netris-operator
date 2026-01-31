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

	"github.com/netrisai/netriswebapi/v2/types/port"
)

func TestNewPortStorage(t *testing.T) {
	storage := NewPortStorage()
	if storage == nil {
		t.Fatal("expected non-nil PortsStorage")
	}
	if storage.Ports != nil {
		t.Errorf("expected Ports to be nil, got %v", storage.Ports)
	}
}

func TestPortsStorage_GetAll_Empty(t *testing.T) {
	storage := NewPortStorage()
	result := storage.GetAll()
	if result != nil {
		t.Errorf("expected nil for empty storage, got %v", result)
	}
}

func TestPortsStorage_GetAll_WithData(t *testing.T) {
	storage := NewPortStorage()
	storage.Ports = []*port.Port{
		{ID: 1, Port_: "eth0", SwitchName: "switch1"},
		{ID: 2, Port_: "eth1", SwitchName: "switch1"},
	}

	result := storage.GetAll()
	if len(result) != 2 {
		t.Errorf("expected 2 ports, got %d", len(result))
	}
}

func TestPortsStorage_FindByName_Found(t *testing.T) {
	storage := NewPortStorage()
	storage.Ports = []*port.Port{
		{ID: 1, Port_: "eth0", SwitchName: "switch1"},
		{ID: 2, Port_: "eth1", SwitchName: "switch1"},
		{ID: 3, Port_: "eth0", SwitchName: "switch2"},
	}

	// FindByName uses format "port@switch"
	result, ok := storage.FindByName("eth1@switch1")
	if !ok {
		t.Fatal("expected to find port 'eth1@switch1'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestPortsStorage_FindByName_NotFound(t *testing.T) {
	storage := NewPortStorage()
	storage.Ports = []*port.Port{
		{ID: 1, Port_: "eth0", SwitchName: "switch1"},
	}

	result, ok := storage.FindByName("eth99@switch99")
	if ok {
		t.Error("expected not to find port 'eth99@switch99'")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestPortsStorage_FindByName_SamePortDifferentSwitch(t *testing.T) {
	storage := NewPortStorage()
	storage.Ports = []*port.Port{
		{ID: 1, Port_: "eth0", SwitchName: "switch1"},
		{ID: 2, Port_: "eth0", SwitchName: "switch2"},
	}

	// Should find the correct port based on switch name
	result, ok := storage.FindByName("eth0@switch2")
	if !ok {
		t.Fatal("expected to find port 'eth0@switch2'")
	}
	if result.ID != 2 {
		t.Errorf("expected ID 2, got %d", result.ID)
	}
}

func TestPortsStorage_FindByID_Found(t *testing.T) {
	storage := NewPortStorage()
	storage.Ports = []*port.Port{
		{ID: 100, Port_: "eth0", SwitchName: "switch1"},
		{ID: 200, Port_: "eth1", SwitchName: "switch1"},
	}

	result, ok := storage.FindByID(200)
	if !ok {
		t.Fatal("expected to find port with ID 200")
	}
	if result.Port_ != "eth1" {
		t.Errorf("expected port 'eth1', got %q", result.Port_)
	}
}

func TestPortsStorage_FindByID_NotFound(t *testing.T) {
	storage := NewPortStorage()
	storage.Ports = []*port.Port{
		{ID: 100, Port_: "eth0", SwitchName: "switch1"},
	}

	// Note: PortsStorage.FindByID doesn't call download on miss (unlike other storage types)
	result, ok := storage.FindByID(999)
	if ok {
		t.Error("expected not to find port with ID 999")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
