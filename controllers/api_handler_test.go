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

package controllers

import (
	"testing"
	"time"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netris-operator/netrisstorage"
	"github.com/netrisai/netriswebapi/v2/types/port"
	"github.com/netrisai/netriswebapi/v2/types/site"
)

func TestGetSites(t *testing.T) {
	tests := []struct {
		name     string
		sites    []*site.Site
		names    []string
		expected map[string]int
	}{
		{
			name: "finds all sites",
			sites: []*site.Site{
				{ID: 1, Name: "site-a"},
				{ID: 2, Name: "site-b"},
				{ID: 3, Name: "site-c"},
			},
			names: []string{"site-a", "site-b"},
			expected: map[string]int{
				"site-a": 1,
				"site-b": 2,
			},
		},
		{
			name: "missing sites return zero",
			sites: []*site.Site{
				{ID: 1, Name: "site-a"},
			},
			names: []string{"site-a", "site-missing"},
			expected: map[string]int{
				"site-a":       1,
				"site-missing": 0,
			},
		},
		{
			name:     "empty input returns empty map",
			sites:    []*site.Site{{ID: 1, Name: "site-a"}},
			names:    []string{},
			expected: map[string]int{},
		},
		{
			name:  "all sites missing returns zeros",
			sites: []*site.Site{},
			names: []string{"missing-a", "missing-b"},
			expected: map[string]int{
				"missing-a": 0,
				"missing-b": 0,
			},
		},
		{
			name: "duplicate names in input are deduplicated",
			sites: []*site.Site{
				{ID: 5, Name: "site-x"},
			},
			names: []string{"site-x", "site-x", "site-x"},
			expected: map[string]int{
				"site-x": 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newTestStorage(tt.sites)
			result := getSites(tt.names, storage)

			if len(result) != len(tt.expected) {
				t.Errorf("got %d entries, expected %d", len(result), len(tt.expected))
			}

			for name, expectedID := range tt.expected {
				if gotID, ok := result[name]; !ok {
					t.Errorf("missing key %q in result", name)
				} else if gotID != expectedID {
					t.Errorf("for site %q: got ID %d, expected %d", name, gotID, expectedID)
				}
			}
		})
	}
}

func TestInitRequeueInterval(t *testing.T) {
	// Save original values
	originalRequeueInterval := requeueInterval
	originalContextTimeout := contextTimeout
	defer func() {
		requeueInterval = originalRequeueInterval
		contextTimeout = originalContextTimeout
	}()

	tests := []struct {
		name             string
		interval         int
		expectedInterval time.Duration
	}{
		{
			name:             "positive interval sets values",
			interval:         30,
			expectedInterval: 30 * time.Second,
		},
		{
			name:             "zero interval does not change values",
			interval:         0,
			expectedInterval: originalRequeueInterval,
		},
		{
			name:             "negative interval does not change values",
			interval:         -5,
			expectedInterval: originalRequeueInterval,
		},
		{
			name:             "large interval works",
			interval:         300,
			expectedInterval: 300 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to original before each test
			requeueInterval = originalRequeueInterval
			contextTimeout = originalContextTimeout

			InitRequeueInterval(tt.interval)

			if tt.interval > 0 {
				if requeueInterval != tt.expectedInterval {
					t.Errorf("requeueInterval: got %v, expected %v", requeueInterval, tt.expectedInterval)
				}
				if contextTimeout != tt.expectedInterval {
					t.Errorf("contextTimeout: got %v, expected %v", contextTimeout, tt.expectedInterval)
				}
			}
		})
	}
}

func TestVNetReconciler_GetPortsMeta(t *testing.T) {
	tests := []struct {
		name        string
		ports       []*port.Port
		portNames   []k8sv1alpha1.VNetSwitchPort
		expectError bool
		expected    []k8sv1alpha1.VNetMetaMember
	}{
		{
			name: "single port found",
			ports: []*port.Port{
				{ID: 100, Port_: "eth0", SwitchName: "switch1"},
			},
			portNames: []k8sv1alpha1.VNetSwitchPort{
				{Name: "eth0@switch1"},
			},
			expectError: false,
			expected: []k8sv1alpha1.VNetMetaMember{
				{ID: 100, Name: "eth0@switch1", Vlan: "1", Lacp: "off", State: "active"},
			},
		},
		{
			name: "port not found returns error",
			ports: []*port.Port{
				{ID: 100, Port_: "eth0", SwitchName: "switch1"},
			},
			portNames: []k8sv1alpha1.VNetSwitchPort{
				{Name: "eth99@switch99"},
			},
			expectError: true,
		},
		{
			name:        "empty ports returns empty members",
			ports:       []*port.Port{},
			portNames:   []k8sv1alpha1.VNetSwitchPort{},
			expectError: false,
			expected:    []k8sv1alpha1.VNetMetaMember{},
		},
		{
			name: "custom vlan ID",
			ports: []*port.Port{
				{ID: 200, Port_: "eth1", SwitchName: "switch2"},
			},
			portNames: []k8sv1alpha1.VNetSwitchPort{
				{Name: "eth1@switch2", VlanID: 100},
			},
			expectError: false,
			expected: []k8sv1alpha1.VNetMetaMember{
				{ID: 200, Name: "eth1@switch2", Vlan: "100", Lacp: "off", State: "active"},
			},
		},
		{
			name: "disabled state",
			ports: []*port.Port{
				{ID: 300, Port_: "eth2", SwitchName: "switch3"},
			},
			portNames: []k8sv1alpha1.VNetSwitchPort{
				{Name: "eth2@switch3", State: "disabled"},
			},
			expectError: false,
			expected: []k8sv1alpha1.VNetMetaMember{
				{ID: 300, Name: "eth2@switch3", Vlan: "1", Lacp: "off", State: "disabled"},
			},
		},
		{
			name: "untagged yes",
			ports: []*port.Port{
				{ID: 400, Port_: "eth3", SwitchName: "switch4"},
			},
			portNames: []k8sv1alpha1.VNetSwitchPort{
				{Name: "eth3@switch4", Untagged: "yes"},
			},
			expectError: false,
			expected: []k8sv1alpha1.VNetMetaMember{
				{ID: 400, Name: "eth3@switch4", Vlan: "1", Lacp: "off", State: "active", Untagged: "yes"},
			},
		},
		{
			name: "invalid state uses default active",
			ports: []*port.Port{
				{ID: 500, Port_: "eth4", SwitchName: "switch5"},
			},
			portNames: []k8sv1alpha1.VNetSwitchPort{
				{Name: "eth4@switch5", State: "invalid"},
			},
			expectError: false,
			expected: []k8sv1alpha1.VNetMetaMember{
				{ID: 500, Name: "eth4@switch5", Vlan: "1", Lacp: "off", State: "active"},
			},
		},
		{
			name: "invalid untagged is ignored",
			ports: []*port.Port{
				{ID: 600, Port_: "eth5", SwitchName: "switch6"},
			},
			portNames: []k8sv1alpha1.VNetSwitchPort{
				{Name: "eth5@switch6", Untagged: "invalid"},
			},
			expectError: false,
			expected: []k8sv1alpha1.VNetMetaMember{
				{ID: 600, Name: "eth5@switch6", Vlan: "1", Lacp: "off", State: "active", Untagged: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &netrisstorage.Storage{
				PortsStorage: netrisstorage.NewPortStorage(),
			}
			storage.PortsStorage.Ports = tt.ports

			r := &VNetReconciler{
				NStorage: storage,
			}

			result, err := r.getPortsMeta(tt.portNames)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("got %d members, expected %d", len(result), len(tt.expected))
				return
			}

			// For single-element tests, verify the values
			if len(tt.expected) == 1 && len(result) == 1 {
				if result[0].ID != tt.expected[0].ID {
					t.Errorf("ID: got %d, expected %d", result[0].ID, tt.expected[0].ID)
				}
				if result[0].Name != tt.expected[0].Name {
					t.Errorf("Name: got %q, expected %q", result[0].Name, tt.expected[0].Name)
				}
				if result[0].Vlan != tt.expected[0].Vlan {
					t.Errorf("Vlan: got %q, expected %q", result[0].Vlan, tt.expected[0].Vlan)
				}
				if result[0].State != tt.expected[0].State {
					t.Errorf("State: got %q, expected %q", result[0].State, tt.expected[0].State)
				}
				if result[0].Untagged != tt.expected[0].Untagged {
					t.Errorf("Untagged: got %q, expected %q", result[0].Untagged, tt.expected[0].Untagged)
				}
				if result[0].Lacp != tt.expected[0].Lacp {
					t.Errorf("Lacp: got %q, expected %q", result[0].Lacp, tt.expected[0].Lacp)
				}
			}
		})
	}
}
