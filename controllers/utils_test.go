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

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netriswebapi/v2/types/dhcp"
)

func TestMakeGateway(t *testing.T) {
	tests := []struct {
		name           string
		gateway        k8sv1alpha1.VNetGateway
		dhcpOptionSets map[string]*dhcp.DHCPOptionSet
		expected       k8sv1alpha1.VNetMetaGateway
	}{
		{
			name: "IPv4 gateway without DHCP",
			gateway: k8sv1alpha1.VNetGateway{
				Prefix: "192.168.1.1/24",
			},
			dhcpOptionSets: nil,
			expected: k8sv1alpha1.VNetMetaGateway{
				Gateway:  "192.168.1.1",
				GwLength: 24,
				Version:  "ipv4",
				DHCP:     false,
			},
		},
		{
			name: "IPv6 gateway without DHCP",
			gateway: k8sv1alpha1.VNetGateway{
				Prefix: "2001:db8::1/64",
			},
			dhcpOptionSets: nil,
			expected: k8sv1alpha1.VNetMetaGateway{
				Gateway:  "2001:db8::1",
				GwLength: 64,
				Version:  "ipv6",
				DHCP:     false,
			},
		},
		{
			name: "IPv4 gateway with DHCP enabled",
			gateway: k8sv1alpha1.VNetGateway{
				Prefix:      "10.0.0.1/24",
				DHCP:        "enabled",
				DHCPStartIP: "10.0.0.100",
				DHCPEndIP:   "10.0.0.200",
			},
			dhcpOptionSets: nil,
			expected: k8sv1alpha1.VNetMetaGateway{
				Gateway:     "10.0.0.1",
				GwLength:    24,
				Version:     "ipv4",
				DHCP:        true,
				DHCPStartIP: "10.0.0.100",
				DHCPEndIP:   "10.0.0.200",
			},
		},
		{
			name: "DHCP with option set",
			gateway: k8sv1alpha1.VNetGateway{
				Prefix:        "172.16.0.1/16",
				DHCP:          "enabled",
				DHCPStartIP:   "172.16.1.1",
				DHCPEndIP:     "172.16.1.254",
				DHCPOptionSet: "custom-options",
			},
			dhcpOptionSets: map[string]*dhcp.DHCPOptionSet{
				"custom-options": {ID: 42, Name: "custom-options"},
			},
			expected: k8sv1alpha1.VNetMetaGateway{
				Gateway:         "172.16.0.1",
				GwLength:        16,
				Version:         "ipv4",
				DHCP:            true,
				DHCPStartIP:     "172.16.1.1",
				DHCPEndIP:       "172.16.1.254",
				DHCPOptionSetID: 42,
			},
		},
		{
			name: "DHCP with missing option set returns zero ID",
			gateway: k8sv1alpha1.VNetGateway{
				Prefix:        "10.10.0.1/24",
				DHCP:          "enabled",
				DHCPOptionSet: "nonexistent",
			},
			dhcpOptionSets: map[string]*dhcp.DHCPOptionSet{},
			expected: k8sv1alpha1.VNetMetaGateway{
				Gateway:         "10.10.0.1",
				GwLength:        24,
				Version:         "ipv4",
				DHCP:            true,
				DHCPOptionSetID: 0,
			},
		},
		{
			name: "invalid CIDR returns empty gateway",
			gateway: k8sv1alpha1.VNetGateway{
				Prefix: "not-a-valid-cidr",
			},
			dhcpOptionSets: nil,
			expected:       k8sv1alpha1.VNetMetaGateway{},
		},
		{
			name: "different prefix lengths",
			gateway: k8sv1alpha1.VNetGateway{
				Prefix: "10.0.0.1/32",
			},
			dhcpOptionSets: nil,
			expected: k8sv1alpha1.VNetMetaGateway{
				Gateway:  "10.0.0.1",
				GwLength: 32,
				Version:  "ipv4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeGateway(tt.gateway, tt.dhcpOptionSets)

			if result.Gateway != tt.expected.Gateway {
				t.Errorf("Gateway: got %q, expected %q", result.Gateway, tt.expected.Gateway)
			}
			if result.GwLength != tt.expected.GwLength {
				t.Errorf("GwLength: got %d, expected %d", result.GwLength, tt.expected.GwLength)
			}
			if result.Version != tt.expected.Version {
				t.Errorf("Version: got %q, expected %q", result.Version, tt.expected.Version)
			}
			if result.DHCP != tt.expected.DHCP {
				t.Errorf("DHCP: got %v, expected %v", result.DHCP, tt.expected.DHCP)
			}
			if result.DHCPStartIP != tt.expected.DHCPStartIP {
				t.Errorf("DHCPStartIP: got %q, expected %q", result.DHCPStartIP, tt.expected.DHCPStartIP)
			}
			if result.DHCPEndIP != tt.expected.DHCPEndIP {
				t.Errorf("DHCPEndIP: got %q, expected %q", result.DHCPEndIP, tt.expected.DHCPEndIP)
			}
			if result.DHCPOptionSetID != tt.expected.DHCPOptionSetID {
				t.Errorf("DHCPOptionSetID: got %d, expected %d", result.DHCPOptionSetID, tt.expected.DHCPOptionSetID)
			}
		})
	}
}

func TestRegParser(t *testing.T) {
	tests := []struct {
		name        string
		valueMatch  []string
		subexpNames []string
		expected    map[string]string
	}{
		{
			name:        "empty inputs",
			valueMatch:  []string{},
			subexpNames: []string{},
			expected:    map[string]string{},
		},
		{
			name:        "single named group",
			valueMatch:  []string{"full-match", "value1"},
			subexpNames: []string{"", "group1"},
			expected:    map[string]string{"group1": "value1"},
		},
		{
			name:        "multiple named groups",
			valueMatch:  []string{"full-match", "val1", "val2", "val3"},
			subexpNames: []string{"", "first", "second", "third"},
			expected:    map[string]string{"first": "val1", "second": "val2", "third": "val3"},
		},
		{
			name:        "unnamed groups are skipped",
			valueMatch:  []string{"full-match", "unnamed", "named-val"},
			subexpNames: []string{"", "", "named"},
			expected:    map[string]string{"named": "named-val"},
		},
		{
			name:        "empty named group value",
			valueMatch:  []string{"full-match", ""},
			subexpNames: []string{"", "empty"},
			expected:    map[string]string{"empty": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := regParser(tt.valueMatch, tt.subexpNames)

			if len(result) != len(tt.expected) {
				t.Errorf("got %d entries, expected %d", len(result), len(tt.expected))
			}

			for key, wantVal := range tt.expected {
				if gotVal, ok := result[key]; !ok {
					t.Errorf("missing key %q", key)
				} else if gotVal != wantVal {
					t.Errorf("for key %q: got %q, expected %q", key, gotVal, wantVal)
				}
			}
		})
	}
}
