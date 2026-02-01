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
	"github.com/netrisai/netriswebapi/v2/types/nat"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNatCompareFieldsForNewMeta(t *testing.T) {
	tests := []struct {
		name     string
		nat      *k8sv1alpha1.Nat
		natMeta  *k8sv1alpha1.NatMeta
		expected bool
	}{
		{
			name: "no changes",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: false,
		},
		{
			name: "generation changed",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "false",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
		{
			name: "import annotation changed",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Annotations: map[string]string{
						"resource.k8s.netris.ai/import":        "true",
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatCRGeneration: 1,
					Imported:        false,
					Reclaim:         false,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := natCompareFieldsForNewMeta(tt.nat, tt.natMeta)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNatMustUpdateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		nat      *k8sv1alpha1.Nat
		expected bool
	}{
		{
			name: "no annotations",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: true,
		},
		{
			name: "valid annotations",
			nat: &k8sv1alpha1.Nat{
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
			name: "missing import annotation",
			nat: &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"resource.k8s.netris.ai/reclaimPolicy": "delete",
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := natMustUpdateAnnotations(tt.nat)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNatUpdateDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name                  string
		annotations           map[string]string
		expectedImport        string
		expectedReclaimPolicy string
	}{
		{
			name:                  "nil annotations get defaults",
			annotations:           nil,
			expectedImport:        "false",
			expectedReclaimPolicy: "delete",
		},
		{
			name:                  "empty annotations get defaults",
			annotations:           map[string]string{},
			expectedImport:        "false",
			expectedReclaimPolicy: "delete",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nat := &k8sv1alpha1.Nat{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}

			natUpdateDefaultAnnotations(nat)

			if nat.GetAnnotations()["resource.k8s.netris.ai/import"] != tt.expectedImport {
				t.Errorf("import: got %q, expected %q",
					nat.GetAnnotations()["resource.k8s.netris.ai/import"], tt.expectedImport)
			}
			if nat.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != tt.expectedReclaimPolicy {
				t.Errorf("reclaimPolicy: got %q, expected %q",
					nat.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"], tt.expectedReclaimPolicy)
			}
		})
	}
}

func TestNatMetaToNetrisUpdate(t *testing.T) {
	tests := []struct {
		name            string
		natMeta         *k8sv1alpha1.NatMeta
		expectedName    string
		expectedAction  string
		expectedSiteID  int
		expectedSrcAddr string
	}{
		{
			name: "basic conversion",
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatName:    "test-nat",
					Action:     "SNAT",
					SiteID:     123,
					SrcAddress: "10.0.0.0/24",
					DstAddress: "192.168.0.0/24",
					Protocol:   "tcp",
					State:      "enabled",
				},
			},
			expectedName:    "test-nat",
			expectedAction:  "SNAT",
			expectedSiteID:  123,
			expectedSrcAddr: "10.0.0.0/24",
		},
		{
			name: "DNAT action",
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatName:    "dnat-rule",
					Action:     "DNAT",
					SiteID:     456,
					SrcAddress: "0.0.0.0/0",
					DstAddress: "203.0.113.1/32",
					DnatToIP:   "10.0.0.100",
					DnatToPort: "8080",
				},
			},
			expectedName:    "dnat-rule",
			expectedAction:  "DNAT",
			expectedSiteID:  456,
			expectedSrcAddr: "0.0.0.0/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NatMetaToNetrisUpdate(tt.natMeta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != tt.expectedName {
				t.Errorf("Name: got %q, expected %q", result.Name, tt.expectedName)
			}
			if result.Action != tt.expectedAction {
				t.Errorf("Action: got %q, expected %q", result.Action, tt.expectedAction)
			}
			if result.Site.ID != tt.expectedSiteID {
				t.Errorf("SiteID: got %d, expected %d", result.Site.ID, tt.expectedSiteID)
			}
			if result.SourceAddress != tt.expectedSrcAddr {
				t.Errorf("SourceAddress: got %q, expected %q", result.SourceAddress, tt.expectedSrcAddr)
			}
		})
	}
}

func TestCompareNatMetaAPIENat(t *testing.T) {
	tests := []struct {
		name     string
		natMeta  *k8sv1alpha1.NatMeta
		apiNat   *nat.NAT
		expected bool
	}{
		{
			name: "all fields match",
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatName:    "test-nat",
					Comment:    "test comment",
					State:      "enabled",
					SiteID:     123,
					Action:     "ACCEPT_SNAT",
					Protocol:   "all",
					SrcAddress: "10.0.0.0/24",
					DstAddress: "192.168.0.0/24",
				},
			},
			apiNat: &nat.NAT{
				Name:    "test-nat",
				Comment: "test comment",
				State: struct {
					Label  string `json:"label"`
					Status string `json:"status"`
					Value  string `json:"value"`
				}{Value: "enabled"},
				Site: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{ID: 123},
				Action: struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Label: "ACCEPT_SNAT"},
				Protocol: struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Value: "all"},
				SourceAddress:      "10.0.0.0/24",
				DestinationAddress: "192.168.0.0/24",
			},
			expected: true,
		},
		{
			name: "name mismatch",
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatName: "nat-a",
				},
			},
			apiNat: &nat.NAT{
				Name: "nat-b",
			},
			expected: false,
		},
		{
			name: "state mismatch",
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatName: "test-nat",
					State:   "enabled",
				},
			},
			apiNat: &nat.NAT{
				Name: "test-nat",
				State: struct {
					Label  string `json:"label"`
					Status string `json:"status"`
					Value  string `json:"value"`
				}{Value: "disabled"},
			},
			expected: false,
		},
		{
			name: "action ACCEPT normalized to ACCEPT_SNAT",
			natMeta: &k8sv1alpha1.NatMeta{
				Spec: k8sv1alpha1.NatMetaSpec{
					NatName:  "test-nat",
					State:    "enabled",
					SiteID:   1,
					Action:   "ACCEPT_SNAT",
					Protocol: "all",
				},
			},
			apiNat: &nat.NAT{
				Name: "test-nat",
				State: struct {
					Label  string `json:"label"`
					Status string `json:"status"`
					Value  string `json:"value"`
				}{Value: "enabled"},
				Site: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{ID: 1},
				Action: struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Label: "ACCEPT"},
				Protocol: struct {
					Label string `json:"label"`
					Value string `json:"value"`
				}{Value: "all"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareNatMetaAPIENat(tt.natMeta, tt.apiNat, newTestLogger())
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}
