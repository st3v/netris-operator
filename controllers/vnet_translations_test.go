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
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFindGatewayDuplicates(t *testing.T) {
	tests := []struct {
		name          string
		gateways      []k8sv1alpha1.VNetGateway
		wantDuplicate string
		wantFound     bool
	}{
		{
			name:          "empty list",
			gateways:      []k8sv1alpha1.VNetGateway{},
			wantDuplicate: "",
			wantFound:     false,
		},
		{
			name: "single gateway",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
			},
			wantDuplicate: "",
			wantFound:     false,
		},
		{
			name: "no duplicates",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
				{Prefix: "10.0.0.1/24"},
				{Prefix: "172.16.0.1/16"},
			},
			wantDuplicate: "",
			wantFound:     false,
		},
		{
			name: "duplicate found",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
				{Prefix: "10.0.0.1/24"},
				{Prefix: "192.168.1.1/24"},
			},
			wantDuplicate: "192.168.1.1/24",
			wantFound:     true,
		},
		{
			name: "returns first duplicate",
			gateways: []k8sv1alpha1.VNetGateway{
				{Prefix: "192.168.1.1/24"},
				{Prefix: "192.168.1.1/24"},
				{Prefix: "10.0.0.1/24"},
				{Prefix: "10.0.0.1/24"},
			},
			wantDuplicate: "192.168.1.1/24",
			wantFound:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDup, gotFound := findGatewayDuplicates(tt.gateways)
			if gotDup != tt.wantDuplicate {
				t.Errorf("duplicate: got %q, want %q", gotDup, tt.wantDuplicate)
			}
			if gotFound != tt.wantFound {
				t.Errorf("found: got %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnetSites(t *testing.T) {
	tests := []struct {
		name      string
		k8sSites  []k8sv1alpha1.VNetMetaSite
		apiSites  []vnet.VNetDetailedSite
		wantMatch bool
	}{
		{
			name:      "both empty",
			k8sSites:  []k8sv1alpha1.VNetMetaSite{},
			apiSites:  []vnet.VNetDetailedSite{},
			wantMatch: true,
		},
		{
			name: "matching sites",
			k8sSites: []k8sv1alpha1.VNetMetaSite{
				{Name: "site-a"},
				{Name: "site-b"},
			},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
				{ID: 2, Name: "site-b"},
			},
			wantMatch: true,
		},
		{
			name: "api has site not in k8s",
			k8sSites: []k8sv1alpha1.VNetMetaSite{
				{Name: "site-a"},
			},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
				{ID: 2, Name: "site-b"},
			},
			wantMatch: false,
		},
		{
			name: "k8s has more sites than api",
			k8sSites: []k8sv1alpha1.VNetMetaSite{
				{Name: "site-a"},
				{Name: "site-b"},
			},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
			},
			wantMatch: true, // function only checks if API sites exist in k8s
		},
		{
			name:     "k8s empty api has sites",
			k8sSites: []k8sv1alpha1.VNetMetaSite{},
			apiSites: []vnet.VNetDetailedSite{
				{ID: 1, Name: "site-a"},
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnetSites(tt.k8sSites, tt.apiSites)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnetTenants(t *testing.T) {
	tests := []struct {
		name       string
		k8sTenants []string
		apiTenants []vnet.VNetDetailedGuestTenant
		wantMatch  bool
	}{
		{
			name:       "both empty",
			k8sTenants: []string{},
			apiTenants: []vnet.VNetDetailedGuestTenant{},
			wantMatch:  true,
		},
		{
			name:       "matching tenants",
			k8sTenants: []string{"tenant-a", "tenant-b"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-a"},
				{ID: 2, Name: "tenant-b"},
			},
			wantMatch: true,
		},
		{
			name:       "different tenants",
			k8sTenants: []string{"tenant-a"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-b"},
			},
			wantMatch: false,
		},
		{
			name:       "different count",
			k8sTenants: []string{"tenant-a", "tenant-b"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-a"},
			},
			wantMatch: false,
		},
		{
			name:       "same tenants different order still matches",
			k8sTenants: []string{"tenant-b", "tenant-a"},
			apiTenants: []vnet.VNetDetailedGuestTenant{
				{ID: 1, Name: "tenant-a"},
				{ID: 2, Name: "tenant-b"},
			},
			wantMatch: true, // diff library treats same elements as matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnetTenants(tt.k8sTenants, tt.apiTenants)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestVnetMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		wantUpdate  bool
	}{
		{
			name:        "no annotations",
			annotations: map[string]string{},
			wantUpdate:  true,
		},
		{
			name: "valid annotations",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			wantUpdate: false,
		},
		{
			name: "import true reclaim retain",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			wantUpdate: false,
		},
		{
			name: "missing import annotation",
			annotations: map[string]string{
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			wantUpdate: true,
		},
		{
			name: "missing reclaimPolicy annotation",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import": "false",
			},
			wantUpdate: true,
		},
		{
			name: "invalid import value",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "invalid",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			wantUpdate: true,
		},
		{
			name: "invalid reclaimPolicy value",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "invalid",
			},
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnet := &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}
			got := vnetMustUpdateAnnotations(vnet)
			if got != tt.wantUpdate {
				t.Errorf("got %v, want %v", got, tt.wantUpdate)
			}
		})
	}
}

func TestVnetUpdateDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name              string
		inputAnnotations  map[string]string
		wantImport        string
		wantReclaimPolicy string
	}{
		{
			name:              "nil annotations get defaults",
			inputAnnotations:  nil,
			wantImport:        "false",
			wantReclaimPolicy: "delete",
		},
		{
			name:              "empty annotations get defaults",
			inputAnnotations:  map[string]string{},
			wantImport:        "false",
			wantReclaimPolicy: "delete",
		},
		{
			name: "preserves import true",
			inputAnnotations: map[string]string{
				"resource.k8s.netris.ai/import": "true",
			},
			wantImport:        "true",
			wantReclaimPolicy: "delete",
		},
		{
			name: "preserves reclaim retain",
			inputAnnotations: map[string]string{
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			wantImport:        "false",
			wantReclaimPolicy: "retain",
		},
		{
			name: "invalid values get overwritten with defaults",
			inputAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "invalid",
				"resource.k8s.netris.ai/reclaimPolicy": "invalid",
			},
			wantImport:        "false",
			wantReclaimPolicy: "delete",
		},
		{
			name: "preserves other annotations",
			inputAnnotations: map[string]string{
				"other-annotation": "value",
			},
			wantImport:        "false",
			wantReclaimPolicy: "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnet := &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.inputAnnotations,
				},
			}
			vnetUpdateDefaultAnnotations(vnet)

			gotImport := vnet.GetAnnotations()["resource.k8s.netris.ai/import"]
			gotReclaim := vnet.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"]

			if gotImport != tt.wantImport {
				t.Errorf("import: got %q, want %q", gotImport, tt.wantImport)
			}
			if gotReclaim != tt.wantReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, want %q", gotReclaim, tt.wantReclaimPolicy)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnetGateways(t *testing.T) {
	tests := []struct {
		name        string
		k8sGateways []k8sv1alpha1.VNetMetaGateway
		apiGateways []vnet.VNetDetailedGateway
		wantMatch   bool
	}{
		{
			name:        "both empty",
			k8sGateways: []k8sv1alpha1.VNetMetaGateway{},
			apiGateways: []vnet.VNetDetailedGateway{},
			wantMatch:   true,
		},
		{
			name: "matching gateways without DHCP",
			k8sGateways: []k8sv1alpha1.VNetMetaGateway{
				{Gateway: "192.168.1.1", GwLength: 24},
			},
			apiGateways: []vnet.VNetDetailedGateway{
				{Prefix: "192.168.1.1/24"},
			},
			wantMatch: true,
		},
		{
			name: "matching gateways with DHCP",
			k8sGateways: []k8sv1alpha1.VNetMetaGateway{
				{
					Gateway:         "192.168.1.1",
					GwLength:        24,
					DHCP:            true,
					DHCPOptionSetID: 1,
					DHCPStartIP:     "192.168.1.100",
					DHCPEndIP:       "192.168.1.200",
				},
			},
			apiGateways: []vnet.VNetDetailedGateway{
				{
					Prefix:      "192.168.1.1/24",
					DHCPEnabled: true,
					DHCP: &vnet.VNetGatewayDHCP{
						OptionSet: vnet.IDName{ID: 1},
						Start:     "192.168.1.100",
						End:       "192.168.1.200",
					},
				},
			},
			wantMatch: true,
		},
		{
			name: "prefix mismatch",
			k8sGateways: []k8sv1alpha1.VNetMetaGateway{
				{Gateway: "192.168.1.1", GwLength: 24},
			},
			apiGateways: []vnet.VNetDetailedGateway{
				{Prefix: "10.0.0.1/24"},
			},
			wantMatch: false,
		},
		{
			name: "DHCP mismatch - k8s has DHCP api does not",
			k8sGateways: []k8sv1alpha1.VNetMetaGateway{
				{
					Gateway:  "192.168.1.1",
					GwLength: 24,
					DHCP:     true,
				},
			},
			apiGateways: []vnet.VNetDetailedGateway{
				{Prefix: "192.168.1.1/24", DHCPEnabled: false},
			},
			wantMatch: false,
		},
		{
			name: "DHCP option set mismatch",
			k8sGateways: []k8sv1alpha1.VNetMetaGateway{
				{
					Gateway:         "192.168.1.1",
					GwLength:        24,
					DHCP:            true,
					DHCPOptionSetID: 1,
				},
			},
			apiGateways: []vnet.VNetDetailedGateway{
				{
					Prefix:      "192.168.1.1/24",
					DHCPEnabled: true,
					DHCP: &vnet.VNetGatewayDHCP{
						OptionSet: vnet.IDName{ID: 2},
					},
				},
			},
			wantMatch: false,
		},
		{
			name: "different gateway count",
			k8sGateways: []k8sv1alpha1.VNetMetaGateway{
				{Gateway: "192.168.1.1", GwLength: 24},
				{Gateway: "10.0.0.1", GwLength: 24},
			},
			apiGateways: []vnet.VNetDetailedGateway{
				{Prefix: "192.168.1.1/24"},
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnetGateways(tt.k8sGateways, tt.apiGateways)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnetMembers(t *testing.T) {
	tests := []struct {
		name       string
		k8sMembers []k8sv1alpha1.VNetMetaMember
		apiMembers []vnet.VNetDetailedPort
		wantMatch  bool
	}{
		{
			name:       "both empty",
			k8sMembers: []k8sv1alpha1.VNetMetaMember{},
			apiMembers: []vnet.VNetDetailedPort{},
			wantMatch:  true,
		},
		{
			name: "matching members",
			k8sMembers: []k8sv1alpha1.VNetMetaMember{
				{ID: 1, Vlan: "100"},
				{ID: 2, Vlan: "100"},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{ID: 1, Vlan: "100"},
				{ID: 2, Vlan: "100"},
			},
			wantMatch: true,
		},
		{
			name: "port ID mismatch",
			k8sMembers: []k8sv1alpha1.VNetMetaMember{
				{ID: 1, Vlan: "100"},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{ID: 2, Vlan: "100"},
			},
			wantMatch: false,
		},
		{
			name: "vlan mismatch",
			k8sMembers: []k8sv1alpha1.VNetMetaMember{
				{ID: 1, Vlan: "100"},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{ID: 1, Vlan: "200"},
			},
			wantMatch: false,
		},
		{
			name: "different count",
			k8sMembers: []k8sv1alpha1.VNetMetaMember{
				{ID: 1, Vlan: "100"},
				{ID: 2, Vlan: "100"},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{ID: 1, Vlan: "100"},
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnetMembers(tt.k8sMembers, tt.apiMembers)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnetMembersUntagged(t *testing.T) {
	tests := []struct {
		name       string
		k8sSpec    k8sv1alpha1.VNetMetaSpec
		apiMembers []vnet.VNetDetailedPort
		wantMatch  bool
	}{
		{
			name: "matching untagged yes",
			k8sSpec: k8sv1alpha1.VNetMetaSpec{
				Members: []k8sv1alpha1.VNetMetaMember{
					{Untagged: "yes"},
				},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{AccessMode: true},
			},
			wantMatch: true,
		},
		{
			name: "matching untagged no with vlan",
			k8sSpec: k8sv1alpha1.VNetMetaSpec{
				VlanID: "100",
				Members: []k8sv1alpha1.VNetMetaMember{
					{Untagged: "no"},
				},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{AccessMode: false},
			},
			wantMatch: true,
		},
		{
			name: "empty untagged with access mode false and vlan set",
			k8sSpec: k8sv1alpha1.VNetMetaSpec{
				VlanID: "100",
				Members: []k8sv1alpha1.VNetMetaMember{
					{Untagged: ""},
				},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{AccessMode: false},
			},
			wantMatch: false,
		},
		{
			name: "empty untagged with access mode true",
			k8sSpec: k8sv1alpha1.VNetMetaSpec{
				Members: []k8sv1alpha1.VNetMetaMember{
					{Untagged: ""},
				},
			},
			apiMembers: []vnet.VNetDetailedPort{
				{AccessMode: true},
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnetMembersUntagged(tt.k8sSpec, tt.apiMembers)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestCompareVNetMetaAPIVnet(t *testing.T) {
	tests := []struct {
		name      string
		vnetMeta  *k8sv1alpha1.VNetMeta
		apiVnet   *vnet.VNetDetailed
		wantMatch bool
	}{
		{
			name: "all fields match",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					Owner:    "admin",
					State:    "active",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-a"},
					},
					Gateways: []k8sv1alpha1.VNetMetaGateway{
						{Gateway: "192.168.1.1", GwLength: 24},
					},
					Members: []k8sv1alpha1.VNetMetaMember{
						{ID: 1, Vlan: "100"},
					},
					Tenants: []string{"tenant-a"},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "test-vnet",
				Tenant: vnet.VNetDetailedTenant{Name: "admin"},
				State:  "active",
				Sites: []vnet.VNetDetailedSite{
					{Name: "site-a"},
				},
				Gateways: []vnet.VNetDetailedGateway{
					{Prefix: "192.168.1.1/24"},
				},
				Ports: []vnet.VNetDetailedPort{
					{ID: 1, Vlan: "100"},
				},
				GuestTenants: []vnet.VNetDetailedGuestTenant{
					{Name: "tenant-a"},
				},
			},
			wantMatch: true,
		},
		{
			name: "name mismatch",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "vnet-a",
					Owner:    "admin",
					State:    "active",
					Sites:    []k8sv1alpha1.VNetMetaSite{},
					Gateways: []k8sv1alpha1.VNetMetaGateway{},
					Members:  []k8sv1alpha1.VNetMetaMember{},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "vnet-b",
				Tenant: vnet.VNetDetailedTenant{Name: "admin"},
				State:  "active",
			},
			wantMatch: false,
		},
		{
			name: "owner mismatch",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					Owner:    "user-a",
					State:    "active",
					Sites:    []k8sv1alpha1.VNetMetaSite{},
					Gateways: []k8sv1alpha1.VNetMetaGateway{},
					Members:  []k8sv1alpha1.VNetMetaMember{},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "test-vnet",
				Tenant: vnet.VNetDetailedTenant{Name: "user-b"},
				State:  "active",
			},
			wantMatch: false,
		},
		{
			name: "state mismatch",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					Owner:    "admin",
					State:    "active",
					Sites:    []k8sv1alpha1.VNetMetaSite{},
					Gateways: []k8sv1alpha1.VNetMetaGateway{},
					Members:  []k8sv1alpha1.VNetMetaMember{},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "test-vnet",
				Tenant: vnet.VNetDetailedTenant{Name: "admin"},
				State:  "disabled",
			},
			wantMatch: false,
		},
		{
			name: "sites mismatch",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					Owner:    "admin",
					State:    "active",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-a"},
					},
					Gateways: []k8sv1alpha1.VNetMetaGateway{},
					Members:  []k8sv1alpha1.VNetMetaMember{},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "test-vnet",
				Tenant: vnet.VNetDetailedTenant{Name: "admin"},
				State:  "active",
				Sites: []vnet.VNetDetailedSite{
					{Name: "site-b"},
				},
			},
			wantMatch: false,
		},
		{
			name: "gateways mismatch",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					Owner:    "admin",
					State:    "active",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-a"},
					},
					Gateways: []k8sv1alpha1.VNetMetaGateway{
						{Gateway: "192.168.1.1", GwLength: 24},
					},
					Members: []k8sv1alpha1.VNetMetaMember{},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "test-vnet",
				Tenant: vnet.VNetDetailedTenant{Name: "admin"},
				State:  "active",
				Sites: []vnet.VNetDetailedSite{
					{Name: "site-a"},
				},
				Gateways: []vnet.VNetDetailedGateway{
					{Prefix: "10.0.0.1/24"},
				},
			},
			wantMatch: false,
		},
		{
			name: "tenants mismatch",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					Owner:    "admin",
					State:    "active",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-a"},
					},
					Gateways: []k8sv1alpha1.VNetMetaGateway{},
					Members:  []k8sv1alpha1.VNetMetaMember{},
					Tenants:  []string{"tenant-a"},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "test-vnet",
				Tenant: vnet.VNetDetailedTenant{Name: "admin"},
				State:  "active",
				Sites: []vnet.VNetDetailedSite{
					{Name: "site-a"},
				},
				GuestTenants: []vnet.VNetDetailedGuestTenant{
					{Name: "tenant-b"},
				},
			},
			wantMatch: false,
		},
		{
			name: "members with auto vlan skip member check",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					Owner:    "admin",
					State:    "active",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-a"},
					},
					Gateways: []k8sv1alpha1.VNetMetaGateway{},
					Members: []k8sv1alpha1.VNetMetaMember{
						{ID: 1, Vlan: "auto"},
					},
				},
			},
			apiVnet: &vnet.VNetDetailed{
				Name:   "test-vnet",
				Tenant: vnet.VNetDetailedTenant{Name: "admin"},
				State:  "active",
				Sites: []vnet.VNetDetailedSite{
					{Name: "site-a"},
				},
				Ports: []vnet.VNetDetailedPort{
					{ID: 999, Vlan: "200"},
				},
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVNetMetaAPIVnet(tt.vnetMeta, tt.apiVnet)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestVnetCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name            string
		vnetGen         int64
		vnetAnnotations map[string]string
		metaGen         int64
		metaImported    bool
		metaReclaim     bool
		wantChanged     bool
	}{
		{
			name:    "no changes",
			vnetGen: 1,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  false,
		},
		{
			name:    "generation changed",
			vnetGen: 2,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  true,
		},
		{
			name:    "import annotation changed",
			vnetGen: 1,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  true,
		},
		{
			name:    "reclaim annotation changed",
			vnetGen: 1,
			vnetAnnotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			metaGen:      1,
			metaImported: false,
			metaReclaim:  false,
			wantChanged:  true,
		},
		{
			name:            "missing annotations treated as false",
			vnetGen:         1,
			vnetAnnotations: map[string]string{},
			metaGen:         1,
			metaImported:    false,
			metaReclaim:     false,
			wantChanged:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnet := &k8sv1alpha1.VNet{
				ObjectMeta: metav1.ObjectMeta{
					Generation:  tt.vnetGen,
					Annotations: tt.vnetAnnotations,
				},
			}
			vnetMeta := &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetCRGeneration: tt.metaGen,
					Imported:         tt.metaImported,
					Reclaim:          tt.metaReclaim,
				},
			}

			got := vnetCompareFieldsForNewMeta(vnet, vnetMeta)
			if got != tt.wantChanged {
				t.Errorf("got %v, want %v", got, tt.wantChanged)
			}
		})
	}
}

func TestVnetMetaToNetrisUpdate(t *testing.T) {
	tests := []struct {
		name             string
		vnetMeta         *k8sv1alpha1.VNetMeta
		expectedName     string
		expectedState    string
		expectedSites    int
		expectedGateways int
		expectedPorts    int
	}{
		{
			name: "basic vnet update",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "test-vnet",
					State:    "active",
					VlanID:   "100",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-1"},
					},
				},
			},
			expectedName:     "test-vnet",
			expectedState:    "active",
			expectedSites:    1,
			expectedGateways: 0,
			expectedPorts:    0,
		},
		{
			name: "vnet with gateways",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "gateway-vnet",
					State:    "active",
					VlanID:   "200",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-1"},
					},
					Gateways: []k8sv1alpha1.VNetMetaGateway{
						{Gateway: "10.0.0.1", GwLength: 24, DHCP: false},
						{Gateway: "10.0.1.1", GwLength: 24, DHCP: true, DHCPStartIP: "10.0.1.10", DHCPEndIP: "10.0.1.100"},
					},
				},
			},
			expectedName:     "gateway-vnet",
			expectedState:    "active",
			expectedSites:    1,
			expectedGateways: 2,
			expectedPorts:    0,
		},
		{
			name: "vnet with ports",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "port-vnet",
					State:    "active",
					VlanID:   "300",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-1"},
					},
					Members: []k8sv1alpha1.VNetMetaMember{
						{Name: "port-1", Vlan: "300", Untagged: "no", ID: 1},
						{Name: "port-2", Vlan: "300", Untagged: "yes", ID: 2},
					},
				},
			},
			expectedName:     "port-vnet",
			expectedState:    "active",
			expectedSites:    1,
			expectedGateways: 0,
			expectedPorts:    2,
		},
		{
			name: "vnet with guest tenants",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "multi-tenant-vnet",
					State:    "active",
					VlanID:   "400",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-1"},
					},
					Tenants: []string{"tenant-a", "tenant-b"},
				},
			},
			expectedName:     "multi-tenant-vnet",
			expectedState:    "active",
			expectedSites:    1,
			expectedGateways: 0,
			expectedPorts:    0,
		},
		{
			name: "vnet with multiple sites",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "multi-site-vnet",
					State:    "active",
					VlanID:   "500",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-1"},
						{Name: "site-2"},
						{Name: "site-3"},
					},
				},
			},
			expectedName:     "multi-site-vnet",
			expectedState:    "active",
			expectedSites:    3,
			expectedGateways: 0,
			expectedPorts:    0,
		},
		{
			name: "vnet with auto vlan",
			vnetMeta: &k8sv1alpha1.VNetMeta{
				Spec: k8sv1alpha1.VNetMetaSpec{
					VnetName: "auto-vlan-vnet",
					State:    "active",
					VlanID:   "auto",
					Sites: []k8sv1alpha1.VNetMetaSite{
						{Name: "site-1"},
					},
				},
			},
			expectedName:     "auto-vlan-vnet",
			expectedState:    "active",
			expectedSites:    1,
			expectedGateways: 0,
			expectedPorts:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VnetMetaToNetrisUpdate(tt.vnetMeta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != tt.expectedName {
				t.Errorf("Name: got %q, expected %q", result.Name, tt.expectedName)
			}
			if result.State != tt.expectedState {
				t.Errorf("State: got %q, expected %q", result.State, tt.expectedState)
			}
			if len(result.Sites) != tt.expectedSites {
				t.Errorf("Sites count: got %d, expected %d", len(result.Sites), tt.expectedSites)
			}
			if len(result.Gateways) != tt.expectedGateways {
				t.Errorf("Gateways count: got %d, expected %d", len(result.Gateways), tt.expectedGateways)
			}
			if len(result.Ports) != tt.expectedPorts {
				t.Errorf("Ports count: got %d, expected %d", len(result.Ports), tt.expectedPorts)
			}
		})
	}
}
