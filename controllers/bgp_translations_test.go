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
	"github.com/netrisai/netriswebapi/v2/types/bgp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBgpCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		bgp      *k8sv1alpha1.BGP
		bgpMeta  *k8sv1alpha1.BGPMeta
		expected bool
	}{
		{
			name: "no changes",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
		{
			name: "reclaim annotation changed",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "retain",
					},
				},
			},
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bgpCompareFieldsForNewMeta(tt.bgp, tt.bgpMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBgpMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		bgp      *k8sv1alpha1.BGP
		expected bool
	}{
		{
			name: "no annotations",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: false,
		},
		{
			name: "import true reclaim retain",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "retain",
					},
				},
			},
			expected: false,
		},
		{
			name: "missing import annotation",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: true,
		},
		{
			name: "missing reclaimPolicy annotation",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import": "false",
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid import value",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "invalid",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: true,
		},
		{
			name: "invalid reclaimPolicy value",
			bgp: &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "invalid",
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bgpMustUpdateAnnotations(tt.bgp)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBgpUpdateDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name                   string
		annotations            map[string]string
		expectedImport         string
		expectedReclaimPolicy  string
	}{
		{
			name:                   "empty annotations get defaults",
			annotations:            map[string]string{},
			expectedImport:         "false",
			expectedReclaimPolicy:  "delete",
		},
		{
			name: "preserves import true",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import": "true",
			},
			expectedImport:        "true",
			expectedReclaimPolicy: "delete",
		},
		{
			name: "preserves reclaim retain",
			annotations: map[string]string{
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			expectedImport:        "false",
			expectedReclaimPolicy: "retain",
		},
		{
			name: "invalid values get overwritten with defaults",
			annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "invalid",
				"resource.k8s.netris.ai/reclaimPolicy": "invalid",
			},
			expectedImport:        "false",
			expectedReclaimPolicy: "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bgp := &k8sv1alpha1.BGP{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			bgpUpdateDefaultAnnotations(bgp)

			if bgp.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					bgp.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if bgp.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					bgp.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestOptionalRouteMapID(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		expectNil   bool
	}{
		{
			name:      "zero returns nil",
			value:     0,
			expectNil: true,
		},
		{
			name:      "negative returns nil",
			value:     -1,
			expectNil: true,
		},
		{
			name:      "positive value returns pointer",
			value:     42,
			expectNil: false,
		},
		{
			name:      "one returns pointer",
			value:     1,
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optionalRouteMapID(tt.value)
			if tt.expectNil {
				if result != nil {
					t.Errorf("expected nil for value %d, got %v", tt.value, *result)
				}
			} else {
				if result == nil {
					t.Errorf("expected non-nil for value %d", tt.value)
				} else if *result != tt.value {
					t.Errorf("got %d, expected %d", *result, tt.value)
				}
			}
		})
	}
}

func TestBGPMetaToNetrisUpdate(t *testing.T) {
	tests := []struct {
		name         string
		bgpMeta      *k8sv1alpha1.BGPMeta
		expectedName string
		expectedSite string
		expectedIP   string
	}{
		{
			name: "basic conversion",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:   "test-bgp",
					Site:      "site-1",
					LocalIP:   "10.0.0.1",
					RemoteIP:  "10.0.0.2",
					NeighborAs: 65000,
					Status:    "enabled",
				},
			},
			expectedName: "test-bgp",
			expectedSite: "site-1",
			expectedIP:   "10.0.0.1",
		},
		{
			name: "with vnet",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:   "bgp-with-vnet",
					Site:      "site-2",
					VnetID:    100,
					LocalIP:   "192.168.1.1",
					NeighborAs: 65001,
				},
			},
			expectedName: "bgp-with-vnet",
			expectedSite: "site-2",
			expectedIP:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BGPMetaToNetrisUpdate(tt.bgpMeta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != tt.expectedName {
				t.Errorf("Name: got %q, expected %q", result.Name, tt.expectedName)
			}
			if result.Site.Name != tt.expectedSite {
				t.Errorf("Site: got %q, expected %q", result.Site.Name, tt.expectedSite)
			}
			if result.LocalIP != tt.expectedIP {
				t.Errorf("LocalIP: got %q, expected %q", result.LocalIP, tt.expectedIP)
			}
		})
	}
}

func TestCompareBGPMetaAPIEBGP(t *testing.T) {
	tests := []struct {
		name      string
		bgpMeta   *k8sv1alpha1.BGPMeta
		apiBGP    *bgp.EBGP
		wantMatch bool
	}{
		{
			name: "all fields match",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					AllowasIn:          2,
					BgpPassword:        "secret",
					Community:          "65000:100",
					Description:        "Test BGP",
					InboundRouteMap:    1,
					IPVersion:          "ipv4",
					LocalIP:            "10.0.0.1",
					LocalPreference:    100,
					Multihop:           3,
					BGPName:            "test-bgp",
					NeighborAddress:    "10.0.0.2",
					NeighborAs:         65000,
					Originate:          "none",
					OutboundRouteMap:   2,
					PrefixLength:       24,
					PrefixLimit:        "500",
					PrefixListInbound:  "in-list",
					PrefixListOutbound: "out-list",
					PrependInbound:     1,
					PrependOutbound:    1,
					RemoteIP:           "10.0.0.3",
					Site:               "site-1",
					Status:             "enabled",
					UpdateSource:       "loopback",
					Vlan:               100,
					Weight:             200,
				},
			},
			apiBGP: &bgp.EBGP{
				AllowasIn:          2,
				BgpPassword:        "secret",
				Community:          "65000:100",
				Description:        "Test BGP",
				InboundRouteMap:    1,
				IPVersion:          "ipv4",
				LocalIP:            "10.0.0.1",
				LocalPreference:    100,
				Multihop:           3,
				Name:               "test-bgp",
				NeighborAddress:    "10.0.0.2",
				NeighborAs:         65000,
				Originate:          "none",
				OutboundRouteMap:   2,
				PrefixLength:       24,
				PrefixLimit:        500,
				PrefixListInbound:  "in-list",
				PrefixListOutbound: "out-list",
				PrependInbound:     1,
				PrependOutbound:    1,
				RemoteIP:           "10.0.0.3",
				SiteName:           "site-1",
				Status:             "enabled",
				UpdateSource:       "loopback",
				Vlan:               100,
				Weight:             200,
			},
			wantMatch: true,
		},
		{
			name: "name mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName: "bgp-a",
				},
			},
			apiBGP: &bgp.EBGP{
				Name: "bgp-b",
			},
			wantMatch: false,
		},
		{
			name: "allowasin mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:   "test-bgp",
					AllowasIn: 1,
				},
			},
			apiBGP: &bgp.EBGP{
				Name:      "test-bgp",
				AllowasIn: 2,
			},
			wantMatch: false,
		},
		{
			name: "community mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:   "test-bgp",
					Community: "65000:100",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:      "test-bgp",
				Community: "65000:200",
			},
			wantMatch: false,
		},
		{
			name: "description mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:     "test-bgp",
					Description: "Description A",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:        "test-bgp",
				Description: "Description B",
			},
			wantMatch: false,
		},
		{
			name: "IP version mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:   "test-bgp",
					IPVersion: "ipv4",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:      "test-bgp",
				IPVersion: "ipv6",
			},
			wantMatch: false,
		},
		{
			name: "local IP mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName: "test-bgp",
					LocalIP: "10.0.0.1",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:    "test-bgp",
				LocalIP: "10.0.0.2",
			},
			wantMatch: false,
		},
		{
			name: "local preference mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:         "test-bgp",
					LocalPreference: 100,
				},
			},
			apiBGP: &bgp.EBGP{
				Name:            "test-bgp",
				LocalPreference: 200,
			},
			wantMatch: false,
		},
		{
			name: "multihop mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:  "test-bgp",
					Multihop: 3,
				},
			},
			apiBGP: &bgp.EBGP{
				Name:     "test-bgp",
				Multihop: 5,
			},
			wantMatch: false,
		},
		{
			name: "neighbor address mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:         "test-bgp",
					NeighborAddress: "10.0.0.1",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:            "test-bgp",
				NeighborAddress: "10.0.0.2",
			},
			wantMatch: false,
		},
		{
			name: "neighbor AS mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:    "test-bgp",
					NeighborAs: 65000,
				},
			},
			apiBGP: &bgp.EBGP{
				Name:       "test-bgp",
				NeighborAs: 65001,
			},
			wantMatch: false,
		},
		{
			name: "remote IP mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:  "test-bgp",
					RemoteIP: "10.0.0.1",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:     "test-bgp",
				RemoteIP: "10.0.0.2",
			},
			wantMatch: false,
		},
		{
			name: "site name mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName: "test-bgp",
					Site:    "site-1",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:     "test-bgp",
				SiteName: "site-2",
			},
			wantMatch: false,
		},
		{
			name: "status mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName: "test-bgp",
					Status:  "enabled",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:   "test-bgp",
				Status: "disabled",
			},
			wantMatch: false,
		},
		{
			name: "weight mismatch",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName: "test-bgp",
					Weight:  100,
				},
			},
			apiBGP: &bgp.EBGP{
				Name:   "test-bgp",
				Weight: 200,
			},
			wantMatch: false,
		},
		{
			name: "vlan mismatch with non-negative-one value",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName: "test-bgp",
					Vlan:    100,
				},
			},
			apiBGP: &bgp.EBGP{
				Name: "test-bgp",
				Vlan: 200,
			},
			wantMatch: false,
		},
		{
			name: "vlan mismatch ignored when spec is -1",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName: "test-bgp",
					Vlan:    -1,
				},
			},
			apiBGP: &bgp.EBGP{
				Name: "test-bgp",
				Vlan: 200,
			},
			wantMatch: true,
		},
		{
			name: "empty neighbor address in meta matches empty in API",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:         "test-bgp",
					NeighborAddress: "",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:            "test-bgp",
				NeighborAddress: "",
			},
			wantMatch: true,
		},
		{
			name: "prefix limit with special case",
			bgpMeta: &k8sv1alpha1.BGPMeta{
				Spec: k8sv1alpha1.BGPMetaSpec{
					BGPName:     "test-bgp",
					PrefixLimit: "0",
				},
			},
			apiBGP: &bgp.EBGP{
				Name:              "test-bgp",
				PrefixLimit:       1000,
				TerminateOnSwitch: "yes",
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := uniReconciler{
				DebugLogger: newTestLogger(),
				NStorage:    newTestStorage(nil),
			}
			got := compareBGPMetaAPIEBGP(tt.bgpMeta, tt.apiBGP, u)
			if got != tt.wantMatch {
				t.Errorf("got %v, want %v", got, tt.wantMatch)
			}
		})
	}
}
